#!/usr/bin/env python3
import csv
import re
import sys
from dataclasses import dataclass
from typing import List


@dataclass
class SummaryLine1:
    elapsed_time: str
    total_vus: str
    complete_iterations: str
    interrupted_iterations: str


@dataclass
class SummaryLine2:
    vus: str
    duration: str


@dataclass
class Title:
    name: str


@dataclass
class Percentage:
    name: str
    percentage: str
    numerator: str
    denominator: str

    def __init__(self, name: str, parts: List[str]):
        self.name = name
        self.percentage = parts[1]
        self.numerator = parts[3]
        self.denominator = parts[5]


@dataclass
class VUS:
    name: str
    val: str
    min: str
    max: str

    def __init__(self, name: str, parts: List[str]):
        self.name = name
        self.val = parts[1]
        for p in parts:
            if p.startswith("min="):
                self.min = p[len("min=")]
            if p.startswith("max="):
                self.max = p[len("max=")]


@dataclass
class Metric:
    name: str
    avg: str
    med: str
    p90: str
    p95: str

    def __init__(self, name: str, parts: List[str]):
        self.name = name
        for p in parts:
            if p.startswith("avg="):
                self.avg = p[len("avg="):]
            if p.startswith("med="):
                self.med = p[len("med="):]
            if p.startswith("p(90)="):
                self.p90 = p[len("p(90)="):]
            if p.startswith("p(95)="):
                self.p95 = p[len("p(95)="):]


@dataclass
class Rate:
    name: str
    total: str
    rate: str

    def __init__(self, name: str, parts: List[str]):
        self.name = name
        self.total = parts[1]
        self.rate = parts[2]


@dataclass
class TestCase:
    title: Title
    http_req_duration: Metric
    http_reqs: Rate
    iteration_duration: Metric
    iterations: Rate
    summary_line_1: SummaryLine1
    summary_line_2: SummaryLine2

    def __init__(self, title):
        self.title = title


def parse_byte_rate(name, line):
    stripped_line = line.rstrip()
    colon = stripped_line.find(":")
    data = line[colon+1:]
    parts = data.split(" ")
    parts = [p for p in parts if p != ""]

    return Rate(
        name=name,
        parts=["", parts[0] + parts[1], parts[2] + parts[3]],
    )


def parse_metric_or_rate(line):
    stripped_line = line.rstrip()
    parts = stripped_line.split(" ")
    parts = [p for p in parts if p != ""]

    if len(parts) < 1:
        return None

    name = parts[0]
    if name.startswith("checks"):
        return Percentage("checks", parts)
    if name.startswith("http_req_failed"):
        return Percentage("http_req_failed", parts)
    if name.startswith("http_req_blocked"):
        return Metric("http_req_blocked", parts)
    if name.startswith("http_req_connecting"):
        return Metric("http_req_connecting", parts)
    if name.startswith("http_req_duration"):
        return Metric("http_req_duration", parts)
    if name.startswith("http_req_receiving"):
        return Metric("http_req_receiving", parts)
    if name.startswith("http_req_sending"):
        return Metric("http_req_sending", parts)
    if name.startswith("http_req_tls_handshaking"):
        return Metric("http_req_tls_handshaking", parts)
    if name.startswith("http_req_waiting"):
        return Metric("http_req_waiting", parts)
    if name.startswith("iteration_duration"):
        return Metric("iteration_duration", parts)
    if name.startswith("http_reqs"):
        return Rate("http_reqs", parts)
    if name.startswith("iterations"):
        return Rate("iterations", parts)
    if name.startswith("data_received"):
        return parse_byte_rate("data_received", line)
    if name.startswith("data_sent"):
        return parse_byte_rate("data_sent", line)
    if name.startswith("{"):
        return Metric("http_req_duration_expected_response_true", parts)
    if name.startswith("vus."):
        return VUS(name="vus", parts=parts)
    if name.startswith("vus_max."):
        return VUS(name="vus_max", parts=parts)


def parse_summary_line_1(line):
    regexp = re.compile(r'running \((.*)\), (\d+)/(\d+) VUs, (\d+) complete and (\d+) interrupted iterations')
    match = regexp.search(line.rstrip())
    if match is None:
        return None

    return SummaryLine1(
        elapsed_time=match.group(1),
        total_vus=match.group(3),
        complete_iterations=match.group(4),
        interrupted_iterations=match.group(5),
    )


def parse_summary_line_2(line):
    regexp = re.compile(r'\[=+\] (\d+) VUs +(\w+)$')
    match = regexp.search(line.rstrip())
    if match is None:
        return None

    return SummaryLine2(
        vus=match.group(1),
        duration=match.group(2),
    )


def parse_title(line):
    stripped_line = line.strip()
    if stripped_line != "":
        return Title(stripped_line)

    return None


def parse_line(line):
    metric_or_rate = parse_metric_or_rate(line)
    if metric_or_rate is not None:
        return metric_or_rate

    summary_line_1 = parse_summary_line_1(line)
    if summary_line_1 is not None:
        return summary_line_1

    summary_line_2 = parse_summary_line_2(line)
    if summary_line_2 is not None:
        return summary_line_2

    title = parse_title(line)
    if title is not None:
        return title


def handle_file(f):
    test_cases = []
    test_case: TestCase | None = None
    for line in f:
        v = parse_line(line)
        if v is not None:
            if isinstance(v, Title):
                test_case = TestCase(v)
                test_cases.append(test_case)
            if isinstance(v, Metric):
                if v.name == "http_req_duration" and test_case is not None:
                    test_case.http_req_duration = v
                if v.name == "iteration_duration" and test_case is not None:
                    test_case.iteration_duration = v
            if isinstance(v, Rate):
                if v.name == "http_reqs" and test_case is not None:
                    test_case.http_reqs = v
                if v.name == "iterations" and test_case is not None:
                    test_case.iterations = v
            if isinstance(v, SummaryLine1) and test_case is not None:
                test_case.summary_line_1 = v
            if isinstance(v, SummaryLine2) and test_case is not None:
                test_case.summary_line_2 = v

    writer = csv.writer(sys.stdout)
    writer.writerow(["case", "duration", "http_req_duration_med", "http_req_duration_p90", "iterations_count", "iterations_rate"])

    for c in test_cases:
        writer.writerow([
            c.title.name,
            c.summary_line_2.duration,
            c.http_req_duration.med,
            c.http_req_duration.p90,
            c.iterations.total,
            c.iterations.rate,
        ])


def main():
    handle_file(sys.stdin)


if __name__ == "__main__":
    main()
