import React, { useCallback, useMemo } from "react";
import cn from "classnames";
import { Button, Select, TextField } from "@radix-ui/themes";
import { FormattedMessage } from "./intl";
import { DateTime } from "luxon";
import styles from "./DateTimePicker.module.css";

export interface DateTimePickerProps {
  className?: string;
  label?: React.ReactElement | null;
  hint?: React.ReactElement | null;
  pickedDateTime: Date | null;
  minDateTime: "now" | null;
  onPickDateTime: (datetime: Date | null) => void;
  showClearButton: boolean;
}

const TIME_INCREMENT_MINUTES = 60;

function getNowWithSecondsStripped(): DateTime {
  return DateTime.now()
    .plus({ minute: 1 })
    .set({
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

function buildTimeOptions(minTime: string | undefined): string[] {
  const options: string[] = [];
  for (
    let minutes = 0;
    minutes < 24 * 60;
    minutes += TIME_INCREMENT_MINUTES
  ) {
    const hour = Math.floor(minutes / 60);
    const minute = minutes % 60;
    const value = `${String(hour).padStart(2, "0")}:${String(minute).padStart(
      2,
      "0"
    )}`;
    if (minTime == null || value >= minTime) {
      options.push(value);
    }
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
    onPickDateTime,
    showClearButton,
  } = props;

  const dateValue = useMemo(
    () => formatDate(pickedDateTime),
    [pickedDateTime]
  );
  const timeValue = useMemo(
    () => formatTime(pickedDateTime),
    [pickedDateTime]
  );
  const minDate = useMemo(
    () =>
      minDateTime === "now"
        ? getNowWithSecondsStripped().toFormat("yyyy-LL-dd")
        : undefined,
    [minDateTime]
  );
  const minTime = useMemo(() => {
    if (minDateTime !== "now" || pickedDateTime == null) {
      return undefined;
    }
    const min = getNowWithSecondsStripped();
    const picked = DateTime.fromJSDate(pickedDateTime);
    return picked.hasSame(min, "day") ? min.toFormat("HH:mm") : undefined;
  }, [minDateTime, pickedDateTime]);

  const timeOptions = useMemo(() => {
    const options = buildTimeOptions(minTime);
    // Keep the currently selected value visible even if it is not on an increment.
    if (timeValue !== "" && !options.includes(timeValue)) {
      return [timeValue, ...options].sort();
    }
    return options;
  }, [minTime, timeValue]);

  const onChangeDate = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      if (event.currentTarget.value === "") {
        onPickDateTime(null);
        return;
      }

      const selectedDate = DateTime.fromISO(event.currentTarget.value);
      if (!selectedDate.isValid) {
        return;
      }

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
      if (minDateTime === "now") {
        const min = getNowWithSecondsStripped();
        if (candidate < min) {
          candidate = min;
        }
      }
      onPickDateTime(candidate.toJSDate());
    },
    [minDateTime, onPickDateTime, pickedDateTime]
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
      if (minDateTime === "now") {
        const min = getNowWithSecondsStripped();
        if (candidate < min) {
          candidate = min;
        }
      }
      onPickDateTime(candidate.toJSDate());
    },
    [minDateTime, onPickDateTime, pickedDateTime]
  );

  const onClickClear = useCallback(() => {
    onPickDateTime(null);
  }, [onPickDateTime]);

  return (
    <div className={cn(className, styles.root)}>
      {label != null ? label : null}
      <div className={styles.inputRow}>
        <TextField.Root
          className={styles.input}
          size="2"
          type="date"
          value={dateValue}
          min={minDate}
          onChange={onChangeDate}
        />
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
