import React, { useCallback, useMemo } from "react";
import cn from "classnames";
import { Button, Select } from "@radix-ui/themes";
import { FormattedMessage } from "./intl";
import { DateTime } from "luxon";
import { DateField } from "./components/v2/DateField/DateField";
import styles from "./DateTimePicker.module.css";

export interface DateTimePickerProps {
  className?: string;
  label?: React.ReactElement | null;
  hint?: React.ReactElement | null;
  pickedDateTime: Date | null;
  minDateTime: "now" | null;
  // Optional upper bound. When set, the picker rejects dates/times after this value.
  maxDateTime?: Date | null;
  onPickDateTime: (datetime: Date | null) => void;
  showClearButton: boolean;
}

const TIME_INCREMENT_MINUTES = 60;

function getNowWithSecondsStripped(): DateTime {
  return DateTime.now().plus({ minute: 1 }).set({
    second: 0,
    millisecond: 0,
  });
}

function formatDate(date: Date | null): string {
  if (date == null) {
    return "";
  }
  return DateTime.fromJSDate(date).toFormat("yyyy-LL-dd");
}

function formatTime(date: Date | null): string {
  if (date == null) {
    return "";
  }
  return DateTime.fromJSDate(date).toFormat("HH:mm");
}

function clampToMax(
  candidate: DateTime,
  maxDateTime: Date | null | undefined
): DateTime {
  if (maxDateTime == null) {
    return candidate;
  }
  const max = DateTime.fromJSDate(maxDateTime);
  if (candidate.valueOf() > max.valueOf()) {
    return max;
  }
  return candidate;
}

function clampToMin(candidate: DateTime, minDateTime: "now" | null): DateTime {
  if (minDateTime !== "now") {
    return candidate;
  }
  const min = getNowWithSecondsStripped();
  if (candidate < min) {
    return min;
  }
  return candidate;
}

function buildTimeOptions(
  minTime: string | undefined,
  maxTime: string | undefined
): string[] {
  const options: string[] = [];
  for (let minutes = 0; minutes < 24 * 60; minutes += TIME_INCREMENT_MINUTES) {
    const hour = Math.floor(minutes / 60);
    const minute = minutes % 60;
    const value = `${String(hour).padStart(2, "0")}:${String(minute).padStart(
      2,
      "0"
    )}`;
    if (minTime != null && value < minTime) {
      continue;
    }
    if (maxTime != null && value > maxTime) {
      continue;
    }
    options.push(value);
  }
  return options;
}

export default function DateTimePicker(
  props: DateTimePickerProps
): React.ReactElement {
  const {
    className,
    label,
    hint,
    pickedDateTime,
    minDateTime,
    maxDateTime = null,
    onPickDateTime,
    showClearButton,
  } = props;

  const dateValue = useMemo(() => formatDate(pickedDateTime), [pickedDateTime]);
  const timeValue = useMemo(() => formatTime(pickedDateTime), [pickedDateTime]);
  const minDate = useMemo(
    () =>
      minDateTime === "now"
        ? getNowWithSecondsStripped().toFormat("yyyy-LL-dd")
        : undefined,
    [minDateTime]
  );
  const maxDate = useMemo(
    () =>
      maxDateTime != null
        ? DateTime.fromJSDate(maxDateTime).toFormat("yyyy-LL-dd")
        : undefined,
    [maxDateTime]
  );
  const minTime = useMemo(() => {
    if (minDateTime !== "now" || pickedDateTime == null) {
      return undefined;
    }
    const min = getNowWithSecondsStripped();
    const picked = DateTime.fromJSDate(pickedDateTime);
    return picked.hasSame(min, "day") ? min.toFormat("HH:mm") : undefined;
  }, [minDateTime, pickedDateTime]);
  const maxTime = useMemo(() => {
    if (maxDateTime == null || pickedDateTime == null) {
      return undefined;
    }
    const max = DateTime.fromJSDate(maxDateTime);
    const picked = DateTime.fromJSDate(pickedDateTime);
    return picked.hasSame(max, "day") ? max.toFormat("HH:mm") : undefined;
  }, [maxDateTime, pickedDateTime]);

  const timeOptions = useMemo(() => {
    const options = buildTimeOptions(minTime, maxTime);
    // Keep the currently selected value visible even if it is not on an increment.
    if (timeValue !== "" && !options.includes(timeValue)) {
      return [timeValue, ...options].sort();
    }
    return options;
  }, [maxTime, minTime, timeValue]);

  const onChangeDate = useCallback(
    (value: string) => {
      if (value === "") {
        onPickDateTime(null);
        return;
      }

      const selectedDate = DateTime.fromISO(value);

      const previous =
        pickedDateTime != null
          ? DateTime.fromJSDate(pickedDateTime)
          : selectedDate.startOf("day");
      let candidate = previous.set({
        year: selectedDate.year,
        month: selectedDate.month,
        day: selectedDate.day,
        second: 0,
        millisecond: 0,
      });
      candidate = clampToMin(candidate, minDateTime);
      candidate = clampToMax(candidate, maxDateTime);
      onPickDateTime(candidate.toJSDate());
    },
    [maxDateTime, minDateTime, onPickDateTime, pickedDateTime]
  );

  const onChangeTime = useCallback(
    (value: string) => {
      if (pickedDateTime == null) {
        return;
      }
      const [hourString, minuteString] = value.split(":");
      const hour = Number(hourString);
      const minute = Number(minuteString);
      if (!Number.isInteger(hour) || !Number.isInteger(minute)) {
        return;
      }

      let candidate = DateTime.fromJSDate(pickedDateTime).set({
        hour,
        minute,
        second: 0,
        millisecond: 0,
      });
      candidate = clampToMin(candidate, minDateTime);
      candidate = clampToMax(candidate, maxDateTime);
      onPickDateTime(candidate.toJSDate());
    },
    [maxDateTime, minDateTime, onPickDateTime, pickedDateTime]
  );

  const onClickClear = useCallback(() => {
    onPickDateTime(null);
  }, [onPickDateTime]);

  return (
    <div className={cn(className, styles.root)}>
      {label != null ? label : null}
      <div className={styles.inputRow}>
        <div className={styles.input}>
          <DateField
            size="2"
            value={dateValue}
            min={minDate}
            max={maxDate}
            onChange={onChangeDate}
            placeholder=""
          />
        </div>
        <div className={styles.timeInput}>
          <Select.Root
            size="2"
            value={timeValue === "" ? undefined : timeValue}
            onValueChange={onChangeTime}
            disabled={pickedDateTime == null}
          >
            <Select.Trigger
              className={styles.timeSelectTrigger}
              variant="surface"
              placeholder="--"
            />
            <Select.Content position="popper">
              {timeOptions.map((option) => (
                <Select.Item key={option} value={option}>
                  {option}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
        </div>
        {showClearButton ? (
          <Button
            type="button"
            size="2"
            variant="outline"
            color="gray"
            disabled={pickedDateTime == null}
            onClick={onClickClear}
          >
            <FormattedMessage id="DateTimePicker.clear" />
          </Button>
        ) : null}
      </div>
      {hint != null ? hint : null}
    </div>
  );
}
