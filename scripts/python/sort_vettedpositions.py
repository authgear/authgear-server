#!/usr/bin/env python3
import sys
import dataclasses
import functools


@dataclasses.dataclass
@functools.total_ordering
class Line:
  path: str
  line: int
  column: int
  contents: str

  def __init__(self, line: str):
    parts = line.split(":")
    if len(parts) < 4:
      raise ValueError(line)
    self.path = parts[0]
    self.line = int(parts[1])
    self.column = int(parts[2])
    self.contents = parts[3].strip()

  def __str__(self) -> str:
    return f"{self.path}:{self.line}:{self.column}: {self.contents}"

  def __lt__(self, other):
    return (self.path, self.line, self.column, self.contents) < (other.path, other.line, other.column, other.contents)

  def __eq__(self, other):
    return (self.path, self.line, self.column, self.contents) == (other.path, other.line, other.column, other.contents)


def main():
  path_to_file = sys.argv[1]
  lines = []
  with open(path_to_file) as f:
    for line in f:
      lines.append(Line(line.strip()))
  lines = sorted(lines)
  with open(path_to_file, "w") as f:
    for line in lines:
      print(line, file=f)

if __name__ == "__main__":
  main()
