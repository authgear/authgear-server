import React, { useState, useCallback, useMemo } from "react";
import cn from "classnames";
import { DatePicker, TimePicker, IComboBox, ITimeRange } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { DateTime, DateObjectUnits } from "luxon";
import DefaultButton from "./DefaultButton";

export interface DateTimePickerProps {
  className?: string;
  label?: React.ReactElement | null;
  hint?: React.ReactElement | null;
  pickedDateTime: Date | null;
  minDateTime: Date | null;
  onPickDateTime: (datetime: Date | null) => void;
  showClearButton: boolean;
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

  const increments = 60;

  // TimePicker has some problem with its controlled component behavior.
  //
  // 1. When we clear the field, value=undefined does not cause it to render empty.
  // 2. Changing the date picker and thus value=something does not cause it to render.
  //
  // So we always remount it in these two cases.
  const [timePickerKey, setTimePickerKey] = useState(0);

  const timeRange: ITimeRange = useMemo(() => {
    // No limit
    if (minDateTime == null || pickedDateTime == null) {
      return {
        start: 0,
        end: 0,
      };
    }

    const startOfDay_minDate = DateTime.fromJSDate(minDateTime).startOf("day");
    const startOfDay_pickedDateTime =
      DateTime.fromJSDate(pickedDateTime).startOf("day");

    // This should not happen.
    if (startOfDay_pickedDateTime.valueOf() < startOfDay_minDate.valueOf()) {
      return {
        start: 0,
        end: 0,
      };
    }
    if (startOfDay_pickedDateTime.valueOf() > startOfDay_minDate.valueOf()) {
      return {
        start: 0,
        end: 0,
      };
    }
    return {
      start: minDateTime.getHours(),
      end: 0,
    };
  }, [minDateTime, pickedDateTime]);

  const onSelectDate_noMinDate = useCallback(
    (date: Date | null | undefined) => {
      if (date == null) {
        onPickDateTime(null);
      } else {
        const datetime =
          pickedDateTime != null ? DateTime.fromJSDate(pickedDateTime) : null;
        onPickDateTime(
          DateTime.fromJSDate(date)
            .set({
              hour: datetime?.hour ?? 0,
              minute: datetime?.minute ?? 0,
              second: 0,
              millisecond: 0,
            })
            .toJSDate()
        );
      }
      setTimePickerKey((prev) => prev + 1);
    },
    [onPickDateTime, pickedDateTime]
  );

  const onSelectDate_withMinDate = useCallback(
    (date: Date | null | undefined) => {
      if (minDateTime == null || pickedDateTime == null || date == null) {
        return;
      }

      const startOfDay_minDate =
        DateTime.fromJSDate(minDateTime).startOf("day");
      const startOfDay_pickedDate = DateTime.fromJSDate(date).startOf("day");

      // Do not allow to pick a date less than minDate.
      if (startOfDay_pickedDate.valueOf() < startOfDay_minDate.valueOf()) {
        return;
      }

      const obj: DateObjectUnits = {
        year: startOfDay_pickedDate.year,
        month: startOfDay_pickedDate.month,
        day: startOfDay_pickedDate.day,
      };

      // Adjust the time.
      if (startOfDay_pickedDate.valueOf() === startOfDay_minDate.valueOf()) {
        const needAdjust =
          pickedDateTime.getHours() < minDateTime.getHours() ||
          (pickedDateTime.getHours() === minDateTime.getHours() &&
            pickedDateTime.getMinutes() < minDateTime.getMinutes());

        if (needAdjust) {
          const d = DateTime.fromJSDate(minDateTime);
          obj.hour = d.hour;
          obj.minute = d.minute;
        }
      }

      onPickDateTime(DateTime.fromJSDate(pickedDateTime).set(obj).toJSDate());
    },
    [minDateTime, onPickDateTime, pickedDateTime]
  );

  const onChange = useCallback(
    (_e: React.FormEvent<IComboBox>, time: Date) => {
      if (pickedDateTime == null) {
        return;
      }
      const datetime = DateTime.fromJSDate(time);
      onPickDateTime(
        DateTime.fromJSDate(pickedDateTime)
          .set({
            hour: datetime.hour,
            minute: datetime.minute,
            second: 0,
            millisecond: 0,
          })
          .toJSDate()
      );
    },
    [onPickDateTime, pickedDateTime]
  );

  const onClickClear = useCallback(() => {
    onPickDateTime(null);
    setTimePickerKey((prev) => prev + 1);
  }, [onPickDateTime]);

  return (
    <div className={cn(className, "flex flex-col")}>
      {label != null ? label : null}
      <div className={"flex flex-row gap-2"}>
        <DatePicker
          className="flex-1"
          value={pickedDateTime ?? undefined}
          onSelectDate={
            minDateTime != null
              ? onSelectDate_withMinDate
              : onSelectDate_noMinDate
          }
          minDate={minDateTime ?? undefined}
        />
        <TimePicker
          key={String(timePickerKey)}
          className="flex-1"
          increments={increments}
          timeRange={timeRange}
          allowFreeform={false}
          showSeconds={false}
          useHour12={false}
          dateAnchor={pickedDateTime ?? undefined}
          value={pickedDateTime ?? undefined}
          onChange={onChange}
        />
        {showClearButton ? (
          <DefaultButton
            className="self-start"
            text={<FormattedMessage id="DateTimePicker.clear" />}
            onClick={onClickClear}
          />
        ) : null}
      </div>
      {hint != null ? hint : null}
    </div>
  );
}
