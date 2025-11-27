import React, { useState, useCallback, useMemo, useContext } from "react";
import cn from "classnames";
import {
  DatePicker,
  TimePicker,
  IComboBox,
  ITimeRange,
  defaultDatePickerStrings,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { DateTime } from "luxon";
import DefaultButton from "./DefaultButton";

export interface DateTimePickerProps {
  className?: string;
  label?: React.ReactElement | null;
  hint?: React.ReactElement | null;
  pickedDateTime: Date | null;
  minDateTime: "now" | null;
  onPickDateTime: (datetime: Date | null) => void;
  showClearButton: boolean;
}

function getNowWithSecondsStripped(): Date {
  return DateTime.fromJSDate(new Date())
    .plus({ minute: 1 })
    .set({
      second: 0,
      millisecond: 0,
    })
    .toJSDate();
}

function formatDate(date?: Date): string {
  if (date == null) {
    return "";
  }
  return DateTime.fromJSDate(date).toFormat("yyyy-LL-dd", {
    locale: "en-US",
  });
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

  const { renderToString } = useContext(Context);

  // TimePicker has some problem with its controlled component behavior.
  //
  // 1. When we clear the field, value=undefined does not cause it to render empty.
  // 2. Changing the date picker and thus value=something does not cause it to render.
  // 3. When allowFreeform=true and an invalid time is input, there is no way to reconcile the input and the selected value.
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

    const min = getNowWithSecondsStripped();

    const startOfDay_minDate = DateTime.fromJSDate(min).startOf("day");
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
      start: min.getHours(),
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

      const min = getNowWithSecondsStripped();

      const pickedDate = DateTime.fromJSDate(date);
      const startOfDay_minDate = DateTime.fromJSDate(min).startOf("day");
      const startOfDay_pickedDate = pickedDate.startOf("day");

      // Do not allow to pick a date less than minDate.
      if (startOfDay_pickedDate.valueOf() < startOfDay_minDate.valueOf()) {
        return;
      }

      let candidate = DateTime.fromObject({
        year: startOfDay_pickedDate.year,
        month: startOfDay_pickedDate.month,
        day: startOfDay_pickedDate.day,
        hour: DateTime.fromJSDate(pickedDateTime).hour,
        minute: DateTime.fromJSDate(pickedDateTime).minute,
        second: 0,
        millisecond: 0,
      });
      if (candidate.toJSDate().getTime() < min.getTime()) {
        candidate = DateTime.fromJSDate(min);
      }

      onPickDateTime(candidate.toJSDate());
      setTimePickerKey((prev) => prev + 1);
    },
    [minDateTime, onPickDateTime, pickedDateTime]
  );

  const onChange = useCallback(
    (_e: React.FormEvent<IComboBox>, time: Date) => {
      if (pickedDateTime == null) {
        return;
      }
      if (!isNaN(time.getTime())) {
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
      }
    },
    [onPickDateTime, pickedDateTime]
  );

  const onClickClear = useCallback(() => {
    onPickDateTime(null);
    setTimePickerKey((prev) => prev + 1);
  }, [onPickDateTime]);

  // DatePicker has poor handling when allowTextInput=true and minDate!=null.
  // 1. It just render isOutOfBoundsErrorMessage, but isOutOfBoundsErrorMessage is undefined by default, so no error message is shown.
  // 2. There is no callback to handle out-of-bound input date, so we have to include our own bound checking in parseDateFromString.
  //
  // The caveat is that we can no longer distinguish between invalid date like "2025-13" or out-of-bound date.
  // https://github.com/microsoft/fluentui/blob/%40fluentui/react_v8.125.1/packages/react/src/components/DatePicker/DatePicker.base.tsx#L158
  const parseDateFromString = useCallback(
    (dateStr: string) => {
      try {
        const dt = DateTime.fromFormat(dateStr, "yyyy-LL-dd", {
          locale: "en-US",
        });
        // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
        if (dt.isValid) {
          const date = dt.toJSDate();
          if (minDateTime != null) {
            const min = getNowWithSecondsStripped();
            const startOfDay_minDate = DateTime.fromJSDate(min).startOf("day");
            const startOfDay_pickedDate =
              DateTime.fromJSDate(date).startOf("day");
            // Do not allow to enter a date less than minDate.
            if (
              startOfDay_pickedDate.valueOf() < startOfDay_minDate.valueOf()
            ) {
              return null;
            }
          }
          return date;
        }
      } catch {}
      return null;
    },
    [minDateTime]
  );

  const onFormatDate = useCallback((date: Date) => {
    return DateTime.fromJSDate(date).toFormat("HH:mm", {
      locale: "en-US",
    });
  }, []);

  const onValidateUserInput = useCallback(
    (timeStr: string) => {
      const check: () => boolean = () => {
        try {
          const timeOnly = DateTime.fromFormat(timeStr, "HH:mm", {
            locale: "en-US",
          });
          // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
          if (timeOnly.isValid) {
            if (minDateTime != null) {
              const min = DateTime.fromJSDate(getNowWithSecondsStripped());
              const dt = timeOnly.set({
                year: min.year,
                month: min.month,
                day: min.day,
              });
              if (dt.valueOf() < min.valueOf()) {
                return false;
              }
            }
            return true;
          }
        } catch {}
        return false;
      };

      const valid = check();
      if (!valid) {
        // Increment the key so that the TimePicker is remounted.
        setTimePickerKey((prev) => prev + 1);
        return "invalid";
      }
      return "";
    },
    [minDateTime]
  );

  const datePickerStrings = useMemo(() => {
    return {
      ...defaultDatePickerStrings,
      isResetStatusMessage: renderToString(
        "DateTimePicker.fluent.isResetStatusMessage"
      ),
    };
  }, [renderToString]);

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
          minDate={minDateTime === "now" ? new Date() : undefined}
          formatDate={formatDate}
          parseDateFromString={parseDateFromString}
          allowTextInput={true}
          strings={datePickerStrings}
        />
        <TimePicker
          key={String(timePickerKey)}
          className="flex-1"
          increments={increments}
          timeRange={timeRange}
          allowFreeform={true}
          showSeconds={false}
          useHour12={false}
          dateAnchor={pickedDateTime ?? undefined}
          value={pickedDateTime ?? undefined}
          onChange={onChange}
          onValidateUserInput={onValidateUserInput}
          onFormatDate={onFormatDate}
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
