#!/usr/bin/env python3
"""
Simulate SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED and
SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED thresholds with mock data.

Formula:
  daily_threshold = max(
    30,                                          # (1) lower bound
    verified_past_14d_rolling_max * 0.2,         # (2) historical baseline
    verified_past_24h * 0.2,                     # (3) daily spike handling
  )
  hourly_threshold = max(
    daily_threshold / 6,
    verified_past_1h * 0.2,                      # (4) hourly spike handling
  )
"""

LOWER_BOUND = 30
HISTORICAL_MULTIPLIER = 0.2
DAILY_SPIKE_MULTIPLIER = 0.2
HOURLY_SPIKE_MULTIPLIER = 0.2


def compute_threshold(
    verified_past_14d_rolling_max: int, verified_past_24h: int, verified_past_1h: int
):
    factor1 = LOWER_BOUND
    factor2 = verified_past_14d_rolling_max * HISTORICAL_MULTIPLIER
    factor3 = verified_past_24h * DAILY_SPIKE_MULTIPLIER
    factor4 = verified_past_1h * HOURLY_SPIKE_MULTIPLIER

    daily = int(max(factor1, factor2, factor3))
    hourly = int(max(daily / 6, factor4))

    binding_map = {factor1: "(1)", factor2: "(2)", factor3: "(3)"}
    daily_binding = binding_map[max(factor1, factor2, factor3)]

    return daily, hourly, daily_binding


scenarios = []


def simulate(
    label,
    verified_past_14d_rolling_max,
    verified_past_24h,
    verified_past_1h,
    unverified_past_24h,
    unverified_past_1h,
):
    daily, hourly, daily_binding = compute_threshold(
        verified_past_14d_rolling_max, verified_past_24h, verified_past_1h
    )

    daily_result = "TRIGGERED" if unverified_past_24h > daily else "ok"
    hourly_result = "TRIGGERED" if unverified_past_1h > hourly else "ok"

    scenarios.append(
        {
            "label": label,
            "14d_max": verified_past_14d_rolling_max,
            "24h_verified": verified_past_24h,
            "1h_verified": verified_past_1h,
            "daily_threshold": daily,
            "daily_binding": daily_binding,
            "hourly_threshold": hourly,
            "24h_unverified": unverified_past_24h,
            "1h_unverified": unverified_past_1h,
            "daily_result": daily_result,
            "hourly_result": hourly_result,
        }
    )

    f1 = LOWER_BOUND
    f2 = verified_past_14d_rolling_max * HISTORICAL_MULTIPLIER
    f3 = verified_past_24h * DAILY_SPIKE_MULTIPLIER
    f4 = verified_past_1h * HOURLY_SPIKE_MULTIPLIER

    print(f"=== {label} ===")
    print(f"  Inputs:")
    print(f"    verified_past_14d_rolling_max : {verified_past_14d_rolling_max:,}")
    print(f"    verified_past_24h             : {verified_past_24h:,}")
    print(f"    verified_past_1h              : {verified_past_1h:,}")
    print(f"    unverified_past_24h           : {unverified_past_24h:,}")
    print(f"    unverified_past_1h            : {unverified_past_1h:,}")
    print(f"  Daily threshold factors:")
    print(f"    (1) lower bound               : {f1}")
    print(
        f"    (2) historical baseline       : {f2:.1f}  ({verified_past_14d_rolling_max:,} * {HISTORICAL_MULTIPLIER})"
    )
    print(
        f"    (3) daily spike               : {f3:.1f}  ({verified_past_24h:,} * {DAILY_SPIKE_MULTIPLIER})"
    )
    print(f"  Daily binding factor            : {daily_binding}")
    print(
        f"  Daily  threshold                : {daily:.1f}  -> {daily_result} (unverified={unverified_past_24h:,})"
    )
    print(f"  Hourly threshold factors:")
    print(f"    daily / 6                     : {daily/6:.1f}")
    print(
        f"    (4) hourly spike              : {f4:.1f}  ({verified_past_1h:,} * {HOURLY_SPIKE_MULTIPLIER})"
    )
    print(
        f"  Hourly threshold                : {hourly:.1f}  -> {hourly_result} (unverified={unverified_past_1h:,})"
    )
    print()


# --- Scenarios (~1k SMS/day project) ---

simulate(
    label="Initial launch (no historical data, ~300 verified in first hour)",
    verified_past_14d_rolling_max=0,
    verified_past_24h=300,
    verified_past_1h=300,
    unverified_past_24h=20,
    unverified_past_1h=20,
)

simulate(
    label="Normal traffic (~1k/day, peak hour ~200)",
    verified_past_14d_rolling_max=1_000,
    verified_past_24h=1_000,
    verified_past_1h=200,
    unverified_past_24h=50,
    unverified_past_1h=10,
)

simulate(
    label="Spike day (~2x normal = 2k/day, peak hour ~400)",
    verified_past_14d_rolling_max=1_000,
    verified_past_24h=2_000,
    verified_past_1h=400,
    unverified_past_24h=150,
    unverified_past_1h=30,
)

simulate(
    label="Attack: quiet day (~1/2 normal = 500/day)",
    verified_past_14d_rolling_max=1_000,
    verified_past_24h=500,
    verified_past_1h=100,
    unverified_past_24h=300,
    unverified_past_1h=60,
)

simulate(
    label="Attack: during spike (~2x normal = 2k/day)",
    verified_past_14d_rolling_max=1_000,
    verified_past_24h=2_000,
    verified_past_1h=400,
    unverified_past_24h=800,
    unverified_past_1h=200,
)

# --- Scenarios (low traffic country, <20 SMS/day) ---

simulate(
    label="[Low traffic] Initial launch (no historical data, ~10 verified in first hour)",
    verified_past_14d_rolling_max=0,
    verified_past_24h=10,
    verified_past_1h=10,
    unverified_past_24h=2,
    unverified_past_1h=2,
)

simulate(
    label="[Low traffic] Normal traffic (~15/day, peak hour ~5)",
    verified_past_14d_rolling_max=15,
    verified_past_24h=12,
    verified_past_1h=5,
    unverified_past_24h=5,
    unverified_past_1h=2,
)

simulate(
    label="[Low traffic] Spike day (~2x normal = 30/day, peak hour ~10)",
    verified_past_14d_rolling_max=15,
    verified_past_24h=30,
    verified_past_1h=10,
    unverified_past_24h=5,
    unverified_past_1h=2,
)

simulate(
    label="[Low traffic] Attack: quiet day (~1/2 normal = 7/day)",
    verified_past_14d_rolling_max=15,
    verified_past_24h=7,
    verified_past_1h=2,
    unverified_past_24h=35,
    unverified_past_1h=8,
)

simulate(
    label="[Low traffic] Attack: during spike (~2x normal = 30/day)",
    verified_past_14d_rolling_max=15,
    verified_past_24h=30,
    verified_past_1h=10,
    unverified_past_24h=50,
    unverified_past_1h=10,
)

# --- Summary markdown tables ---

print("=" * 80)
print("SUMMARY")
print("=" * 80)
print()


def fmt(v):
    return (
        f"{v:,}"
        if isinstance(v, int)
        else (f"{v:.1f}" if isinstance(v, float) else str(v))
    )


headers = [
    "Scenario",
    "Daily threshold",
    "Daily",
    "Hourly threshold",
    "Hourly",
]

col_keys = [
    "label",
    "daily_threshold",
    "daily_result",
    "hourly_threshold",
    "hourly_result",
]


def print_table(title, table_scenarios):
    rows = [[fmt(s[k]) for k in col_keys] for s in table_scenarios]
    col_widths = [
        max(len(h), max(len(r[i]) for r in rows)) for i, h in enumerate(headers)
    ]

    def row_str(cells):
        return "| " + " | ".join(c.ljust(w) for c, w in zip(cells, col_widths)) + " |"

    separator = "| " + " | ".join("-" * w for w in col_widths) + " |"

    print(f"### {title}\n")
    print(row_str(headers))
    print(separator)
    for row in rows:
        print(row_str(row))
    print()


print_table("~1k SMS/day", scenarios[:5])
print_table("Low traffic country (<20 SMS/day)", scenarios[5:])
