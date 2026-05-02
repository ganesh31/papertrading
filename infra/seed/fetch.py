#!/usr/bin/env python3
"""
One-off market data fetcher for Phase 1 (equity-first).

  python fetch.py minute --symbols=INFY,RELIANCE --days=7
  python fetch.py bhavcopy --from=2024-01-15 --to=2024-01-18

Requires DATABASE_URL (or --database-url). Loads repo-root .env when present.

NSE bhavcopy:
  - Dates before 2024-07-09: legacy cm*.zip from archives.nseindia.com (classic CSV).
  - On/after 2024-07-09: CM-UDiFF-style CSV from a configurable mirror (default: chartiny/nse-cm-bhavcopy
    on GitHub). Override base with PT_NSE_CM_BHAVCOPY_BASE or --udiff-base (official NSE zip paths
    often need browser cookies; mirror is for dev convenience).
"""

from __future__ import annotations

import argparse
import calendar
import io
import os
import sys
import zipfile
from datetime import date, datetime, timedelta
from pathlib import Path
from typing import Any, Iterable
from zoneinfo import ZoneInfo

import pandas as pd
import psycopg
import requests
import yfinance as yf
from dotenv import load_dotenv

REPO_ROOT = Path(__file__).resolve().parents[2]
IST = ZoneInfo("Asia/Kolkata")
UTC = ZoneInfo("UTC")

# NSE discontinued classic CM bhav zip in favour of UDiFF (see NSE circular ~July 2024).
LEGACY_BHAV_LAST = date(2024, 7, 8)

DEFAULT_UDIFF_BASE = os.environ.get(
    "PT_NSE_CM_BHAVCOPY_BASE",
    "https://raw.githubusercontent.com/chartiny/nse-cm-bhavcopy/master",
)

HTTP_TIMEOUT = 120
NSE_HEADERS = {
    "User-Agent": "Mozilla/5.0 (compatible; papertrading-fetch/1.0; +https://github.com/ganesh/papertrading)",
    "Accept": "*/*",
}


def _load_env() -> None:
    env_path = REPO_ROOT / ".env"
    if env_path.is_file():
        load_dotenv(env_path)


def _dsn(cli_url: str | None) -> str:
    dsn = cli_url or os.environ.get("DATABASE_URL", "")
    if not dsn.strip():
        sys.exit("DATABASE_URL is not set (or pass --database-url)")
    return dsn


def _yahoo_symbol(base: str) -> str:
    b = base.strip().upper()
    if b.endswith(".NS"):
        return b
    if b.endswith("-EQ"):
        b = b[:-3]
    return f"{b}.NS"


def _tradingsymbol(base: str) -> str:
    b = base.strip().upper()
    if b.endswith("-EQ"):
        return b
    if b.endswith(".NS"):
        b = b[:-3]
    return f"{b}-EQ"


def _connect(dsn: str) -> psycopg.Connection:
    return psycopg.connect(dsn)


def _resolve_instrument_ids(
    conn: psycopg.Connection, tradingsymbols: list[str]
) -> dict[str, str]:
    if not tradingsymbols:
        return {}
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT tradingsymbol, instrument_id
            FROM ref.instruments
            WHERE exchange = 'NSE' AND segment = 'NSE_EQ' AND tradingsymbol = ANY(%s)
            """,
            (tradingsymbols,),
        )
        rows = cur.fetchall()
    return {str(r[0]): str(r[1]) for r in rows}


def _normalize_ohlcv_df(df: pd.DataFrame) -> pd.DataFrame:
    if df.empty:
        return df
    df = df.copy()
    df.columns = [str(c).strip().lower() for c in df.columns]
    keep = [c for c in ("open", "high", "low", "close", "volume") if c in df.columns]
    if len(keep) < 5:
        raise ValueError(f"unexpected OHLCV columns: {list(df.columns)}")
    out = df[keep].copy()
    out.columns = ["open", "high", "low", "close", "volume"]
    for c in ("open", "high", "low", "close"):
        out[c] = pd.to_numeric(out[c], errors="coerce")
    out["volume"] = pd.to_numeric(out["volume"], errors="coerce").fillna(0).astype("int64")
    return out.dropna(subset=["open", "high", "low", "close"], how="any")


def _to_ist_timestamp_index(idx: pd.DatetimeIndex) -> pd.DatetimeIndex:
    if idx.tz is None:
        idx = idx.tz_localize(UTC)
    else:
        idx = idx.tz_convert(UTC)
    return idx.tz_convert(IST)


def cmd_minute(args: argparse.Namespace) -> None:
    days = int(args.days)
    if days < 1 or days > 7:
        sys.exit("minute: --days must be between 1 and 7 for yfinance 1m data")

    bases = [s.strip() for s in args.symbols.split(",") if s.strip()]
    if not bases:
        sys.exit("minute: --symbols is required")

    tradingsymbols = [_tradingsymbol(b) for b in bases]
    dsn = _dsn(args.database_url)

    with _connect(dsn) as conn:
        sym_map = _resolve_instrument_ids(conn, tradingsymbols)
        missing = [s for s in tradingsymbols if s not in sym_map]
        if missing:
            sys.exit(
                "minute: unknown tradingsymbols in ref.instruments (run pt instruments sync first): "
                + ", ".join(missing)
            )

        upsert_sql = """
            INSERT INTO md.bars_1m (instrument_id, ts, open, high, low, close, volume, source)
            VALUES (%s, %s, %s, %s, %s, %s, %s, 'yfinance')
            ON CONFLICT (instrument_id, ts) DO UPDATE SET
              open = EXCLUDED.open,
              high = EXCLUDED.high,
              low = EXCLUDED.low,
              close = EXCLUDED.close,
              volume = EXCLUDED.volume,
              source = EXCLUDED.source
        """

        total = 0
        for base in bases:
            ysym = _yahoo_symbol(base)
            iid = sym_map[_tradingsymbol(base)]
            print(f"minute: fetching {ysym} …", flush=True)
            t = yf.Ticker(ysym)
            raw = t.history(interval="1m", period=f"{days}d", auto_adjust=False)
            if raw.empty:
                print(f"minute: warning: no rows for {ysym}", flush=True)
                continue
            raw = _normalize_ohlcv_df(raw)
            raw.index = _to_ist_timestamp_index(pd.DatetimeIndex(raw.index))

            batch: list[tuple[Any, ...]] = []
            for ts, row in raw.iterrows():
                ts_dt = ts.to_pydatetime()
                batch.append(
                    (
                        iid,
                        ts_dt,
                        float(row["open"]),
                        float(row["high"]),
                        float(row["low"]),
                        float(row["close"]),
                        int(row["volume"]),
                    )
                )
            with conn.cursor() as cur:
                cur.executemany(upsert_sql, batch)
            conn.commit()
            total += len(batch)
            print(f"minute: upserted {len(batch)} bars for {ysym}", flush=True)

        print(f"minute: done ({total} bars total)", flush=True)


def _legacy_bhav_zip_url(d: date) -> str:
    mon = calendar.month_abbr[d.month].upper()
    return (
        f"https://archives.nseindia.com/content/historical/EQUITIES/"
        f"{d.year}/{mon}/cm{d.day:02d}{mon}{d.year}bhav.csv.zip"
    )


def _udiff_bhav_csv_url(d: date, base: str) -> str:
    b = base.rstrip("/")
    return f"{b}/{d.year}/nse-cm-bhavcopy-{d.isoformat()}.csv"


def _http_get(url: str) -> requests.Response:
    return requests.get(url, headers=NSE_HEADERS, timeout=HTTP_TIMEOUT)


def _download_legacy_bhav(d: date, dest_dir: Path) -> Path | None:
    url = _legacy_bhav_zip_url(d)
    r = _http_get(url)
    if r.status_code != 200:
        return None
    dest_dir.mkdir(parents=True, exist_ok=True)
    zpath = dest_dir / f"cm{d:%d}{calendar.month_abbr[d.month].upper()}{d.year}bhav.csv.zip"
    zpath.write_bytes(r.content)
    with zipfile.ZipFile(io.BytesIO(r.content)) as zf:
        names = zf.namelist()
        if not names:
            return None
        inner = names[0]
        csv_bytes = zf.read(inner)
    out_csv = dest_dir / inner
    out_csv.write_bytes(csv_bytes)
    return out_csv


def _download_udiff_bhav(d: date, base: str, dest_dir: Path) -> Path | None:
    url = _udiff_bhav_csv_url(d, base)
    r = _http_get(url)
    if r.status_code != 200:
        return None
    dest_dir.mkdir(parents=True, exist_ok=True)
    out = dest_dir / f"nse-cm-bhavcopy-{d.isoformat()}.csv"
    out.write_bytes(r.content)
    return out


def _parse_legacy_bhav_csv(path: Path, trade_date: date) -> pd.DataFrame:
    df = pd.read_csv(path)
    df.columns = [str(c).strip() for c in df.columns]
    if "SERIES" not in df.columns or "SYMBOL" not in df.columns:
        raise ValueError(f"unexpected legacy bhav columns: {list(df.columns)}")
    eq = df[df["SERIES"].astype(str).str.upper() == "EQ"].copy()
    eq["trade_date"] = trade_date
    eq["tradingsymbol"] = eq["SYMBOL"].astype(str).str.strip() + "-EQ"
    return eq


def _parse_udiff_bhav_csv(path: Path) -> pd.DataFrame:
    df = pd.read_csv(path)
    df.columns = [str(c).strip() for c in df.columns]
    need = {"TradDt", "Sgmt", "FinInstrmTp", "SctySrs", "TckrSymb"}
    if not need.issubset(set(df.columns)):
        raise ValueError(f"unexpected UDiFF bhav columns: {list(df.columns)}")
    m = (
        (df["Sgmt"].astype(str).str.upper() == "CM")
        & (df["FinInstrmTp"].astype(str).str.upper() == "STK")
        & (df["SctySrs"].astype(str).str.upper() == "EQ")
    )
    eq = df.loc[m].copy()
    eq["trade_date"] = pd.to_datetime(eq["TradDt"]).dt.date
    eq["tradingsymbol"] = eq["TckrSymb"].astype(str).str.strip() + "-EQ"
    return eq


def _upsert_bhav_rows(conn: psycopg.Connection, frame: pd.DataFrame, *, udiff: bool) -> int:
    if frame.empty:
        return 0

    sym_col = "tradingsymbol"
    tradingsymbols = sorted(frame[sym_col].astype(str).unique().tolist())
    idmap = _resolve_instrument_ids(conn, tradingsymbols)
    rows: list[tuple[Any, ...]] = []

    for _, r in frame.iterrows():
        tsym = str(r[sym_col])
        iid = idmap.get(tsym)
        if not iid:
            continue
        td = r["trade_date"]
        if isinstance(td, pd.Timestamp):
            trade_date = td.date()
        elif isinstance(td, datetime):
            trade_date = td.date()
        else:
            trade_date = td  # type: ignore[assignment]

        if udiff:
            o = pd.to_numeric(r["OpnPric"], errors="coerce")
            h = pd.to_numeric(r["HghPric"], errors="coerce")
            lo = pd.to_numeric(r["LwPric"], errors="coerce")
            c = pd.to_numeric(r["ClsPric"], errors="coerce")
            la = pd.to_numeric(r["LastPric"], errors="coerce")
            pc = pd.to_numeric(r["PrvsClsgPric"], errors="coerce")
            vol = pd.to_numeric(r["TtlTradgVol"], errors="coerce")
            tov = pd.to_numeric(r["TtlTrfVal"], errors="coerce")
        else:
            o = pd.to_numeric(r["OPEN"], errors="coerce")
            h = pd.to_numeric(r["HIGH"], errors="coerce")
            lo = pd.to_numeric(r["LOW"], errors="coerce")
            c = pd.to_numeric(r["CLOSE"], errors="coerce")
            la = pd.to_numeric(r["LAST"], errors="coerce")
            pc = pd.to_numeric(r["PREVCLOSE"], errors="coerce")
            vol = pd.to_numeric(r["TOTTRDQTY"], errors="coerce")
            tov = pd.to_numeric(r["TOTTRDVAL"], errors="coerce")

        if any(pd.isna(x) for x in (o, h, lo, c, la, pc, vol, tov)):
            continue
        rows.append(
            (
                iid,
                trade_date,
                float(o),
                float(h),
                float(lo),
                float(c),
                float(la),
                float(pc),
                int(vol),
                float(tov),
            )
        )

    if not rows:
        return 0

    sql = """
        INSERT INTO md.bhav_eq (
          instrument_id, trade_date, open, high, low, close, last, prev_close, volume, turnover
        ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
        ON CONFLICT (instrument_id, trade_date) DO UPDATE SET
          open = EXCLUDED.open,
          high = EXCLUDED.high,
          low = EXCLUDED.low,
          close = EXCLUDED.close,
          last = EXCLUDED.last,
          prev_close = EXCLUDED.prev_close,
          volume = EXCLUDED.volume,
          turnover = EXCLUDED.turnover
    """
    with conn.cursor() as cur:
        cur.executemany(sql, rows)
    conn.commit()
    return len(rows)


def _daterange(d0: date, d1: date) -> Iterable[date]:
    d = d0
    while d <= d1:
        yield d
        d += timedelta(days=1)


def cmd_bhavcopy(args: argparse.Namespace) -> None:
    d0 = datetime.strptime(args.date_from, "%Y-%m-%d").date()
    d1 = datetime.strptime(args.date_to, "%Y-%m-%d").date()
    if d0 > d1:
        sys.exit("bhavcopy: --from must be <= --to")

    udiff_base = (args.udiff_base or DEFAULT_UDIFF_BASE).rstrip("/")
    root = Path(args.out_dir)
    dsn = _dsn(args.database_url)

    with _connect(dsn) as conn:
        total_rows = 0
        for d in _daterange(d0, d1):
            day_dir = root / d.isoformat()
            csv_path: Path | None = None
            udiff_mode = False
            if d <= LEGACY_BHAV_LAST:
                csv_path = _download_legacy_bhav(d, day_dir)
            if csv_path is None:
                csv_path = _download_udiff_bhav(d, udiff_base, day_dir)
                if csv_path is not None:
                    udiff_mode = True
            if csv_path is None:
                print(f"bhavcopy: skip {d} (no file; likely holiday/weekend or 404)", flush=True)
                continue

            frame = _parse_udiff_bhav_csv(csv_path) if udiff_mode else _parse_legacy_bhav_csv(csv_path, d)
            n = _upsert_bhav_rows(conn, frame, udiff=udiff_mode)
            total_rows += n
            print(f"bhavcopy: {d} imported {n} EQ rows into md.bhav_eq", flush=True)

        print(f"bhavcopy: done ({total_rows} row writes)", flush=True)


def _add_db_url(p: argparse.ArgumentParser) -> None:
    p.add_argument(
        "--database-url",
        default=os.environ.get("DATABASE_URL", ""),
        help="Postgres URL (default: $DATABASE_URL)",
    )


def main() -> None:
    _load_env()
    p = argparse.ArgumentParser(description="Paper trading seed data fetcher")

    sub = p.add_subparsers(dest="cmd", required=True)

    pm = sub.add_parser("minute", help="1m bars from yfinance → md.bars_1m")
    _add_db_url(pm)
    pm.add_argument("--symbols", required=True, help="Comma-separated bases, e.g. INFY,RELIANCE")
    pm.add_argument("--days", type=int, default=7, help="Lookback days (1–7 for 1m)")
    pm.set_defaults(func=cmd_minute)

    pb = sub.add_parser("bhavcopy", help="NSE equity EOD bhav → md.bhav_eq")
    _add_db_url(pb)
    pb.add_argument("--from", dest="date_from", required=True, help="YYYY-MM-DD (inclusive)")
    pb.add_argument("--to", dest="date_to", required=True, help="YYYY-MM-DD (inclusive)")
    pb.add_argument(
        "--out-dir",
        default=str(REPO_ROOT / "infra" / "seed" / "bhavcopy"),
        help="Download/extract directory (default: infra/seed/bhavcopy)",
    )
    pb.add_argument(
        "--udiff-base",
        default="",
        help="Override PT_NSE_CM_BHAVCOPY_BASE for UDiFF CSV mirror (dates after legacy zip era)",
    )
    pb.set_defaults(func=cmd_bhavcopy)

    args = p.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
