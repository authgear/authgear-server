import React, { useCallback, useMemo, useRef } from "react";
import { CalendarIcon } from "@radix-ui/react-icons";
import { DateTime } from "luxon";
import { TextField } from "../TextField/TextField";
import styles from "./DateField.module.css";

const DATE_FORMAT = "yyyy-MM-dd";
const PLACEHOLDER = "yyyy-MM-dd";

function toDateString(date: Date | null | undefined): string {
  if (date == null || isNaN(date.getTime())) {
    return "";
  }
  return DateTime.fromJSDate(date).toFormat(DATE_FORMAT);
}

function parseDateString(value: string): Date | null {
  if (value === "") {
    return null;
  }
  const datetime = DateTime.fromFormat(value, DATE_FORMAT);
  return datetime.toJSDate();
}

export interface DateFieldProps {
  size?: "2" | "3";
  label?: React.ReactNode;
  /** Optional label typography size; defaults to `size` when omitted. */
  labelSize?: "2" | "3";
  disabled?: boolean;
  /** yyyy-MM-dd string, or empty string when unset. */
  value: string;
  onChange: (value: string) => void;
  min?: string;
  max?: string;
  placeholder?: string;
  /** Accessible name for the native date picker control. */
  ariaLabel?: string;
}

/**
 * Date-only field that always displays `yyyy-MM-dd`, keeps a calendar icon,
 * and opens the native date picker when the input (or icon) is clicked.
 */
export function DateField(props: DateFieldProps): React.ReactElement {
  const {
    size = "2",
    label,
    labelSize,
    disabled,
    value,
    onChange,
    min,
    max,
    placeholder = PLACEHOLDER,
    ariaLabel,
  } = props;

  const pickerRef = useRef<HTMLInputElement>(null);

  // Native date inputs only accept a valid yyyy-MM-dd value.
  const pickerValue = useMemo(() => {
    if (value === "") {
      return "";
    }
    return parseDateString(value) != null ? value : "";
  }, [value]);

  const openPicker = useCallback(() => {
    if (disabled) {
      return;
    }
    const input = pickerRef.current;
    if (input == null) {
      return;
    }
    try {
      if (typeof input.showPicker === "function") {
        input.showPicker();
      } else {
        input.click();
      }
    } catch {
      input.click();
    }
  }, [disabled]);

  const onChangeText = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      onChange(event.target.value);
    },
    [onChange]
  );

  const onPickDate = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      // input[type=date] always yields yyyy-MM-dd (or "").
      onChange(event.target.value);
    },
    [onChange]
  );

  const pickerAriaLabel =
    ariaLabel ?? (typeof label === "string" ? label : placeholder);

  return (
    <TextField
      size={size}
      labelSize={labelSize}
      type="text"
      label={label}
      value={value}
      placeholder={placeholder}
      onChange={onChangeText}
      onClick={openPicker}
      disabled={disabled}
      inputClassName={styles.root}
      suffixPlain={true}
      suffix={
        <span className={styles.pickerWrap} onClick={openPicker}>
          <CalendarIcon
            className={styles.pickerIcon}
            width="1rem"
            height="1rem"
            aria-hidden={true}
          />
          <input
            ref={pickerRef}
            type="date"
            className={styles.nativePicker}
            value={pickerValue}
            min={min}
            max={max}
            onChange={onPickDate}
            disabled={disabled}
            tabIndex={-1}
            aria-label={pickerAriaLabel}
          />
        </span>
      }
    />
  );
}

export interface DateFieldDateProps
  extends Omit<DateFieldProps, "value" | "onChange"> {
  value: Date | null | undefined;
  onChange: (date: Date | null) => void;
}

/**
 * DateField variant that speaks in `Date | null` instead of yyyy-MM-dd strings.
 */
export function DateFieldDate(props: DateFieldDateProps): React.ReactElement {
  const { value, onChange, min, max, ...rest } = props;

  const stringValue = useMemo(() => toDateString(value), [value]);
  const minString = useMemo(() => (min != null ? min : undefined), [min]);
  const maxString = useMemo(() => (max != null ? max : undefined), [max]);

  const onChangeString = useCallback(
    (next: string) => {
      onChange(parseDateString(next));
    },
    [onChange]
  );

  return (
    <DateField
      {...rest}
      value={stringValue}
      onChange={onChangeString}
      min={minString}
      max={maxString}
    />
  );
}

// ── Date + time variant ───────────────────────────────────────────────
// The displayed text is `yyyy-MM-dd HH:mm`; the native picker
// (input[type=datetime-local]) speaks `yyyy-MM-dd'T'HH:mm`.

const DATETIME_DISPLAY_FORMAT = "yyyy-MM-dd HH:mm";
const DATETIME_PICKER_FORMAT = "yyyy-MM-dd'T'HH:mm";
const DATETIME_PLACEHOLDER = "yyyy-MM-dd HH:mm";

function toDateTimeString(date: Date | null | undefined): string {
  if (date == null || isNaN(date.getTime())) {
    return "";
  }
  return DateTime.fromJSDate(date).toFormat(DATETIME_DISPLAY_FORMAT);
}

function parseDateTimeString(value: string): Date | null {
  if (value === "") {
    return null;
  }
  const date = DateTime.fromFormat(value, DATETIME_DISPLAY_FORMAT).toJSDate();
  return isNaN(date.getTime()) ? null : date;
}

function toPickerString(date: Date | null | undefined): string {
  if (date == null || isNaN(date.getTime())) {
    return "";
  }
  return DateTime.fromJSDate(date).toFormat(DATETIME_PICKER_FORMAT);
}

export interface DateTimeFieldProps {
  size?: "2" | "3";
  label?: React.ReactNode;
  labelSize?: "2" | "3";
  disabled?: boolean;
  /** `yyyy-MM-dd HH:mm` string, or empty string when unset. */
  value: string;
  onChange: (value: string) => void;
  /** `yyyy-MM-dd HH:mm` bounds for the native picker. */
  min?: string;
  max?: string;
  placeholder?: string;
  ariaLabel?: string;
}

/**
 * Date + time field that displays `yyyy-MM-dd HH:mm` and opens the native
 * datetime picker when the input (or icon) is clicked.
 */
export function DateTimeField(props: DateTimeFieldProps): React.ReactElement {
  const {
    size = "2",
    label,
    labelSize,
    disabled,
    value,
    onChange,
    min,
    max,
    placeholder = DATETIME_PLACEHOLDER,
    ariaLabel,
  } = props;

  const pickerRef = useRef<HTMLInputElement>(null);

  const pickerValue = useMemo(
    () => toPickerString(parseDateTimeString(value)),
    [value]
  );
  const pickerMin = useMemo(
    () => toPickerString(parseDateTimeString(min ?? "")),
    [min]
  );
  const pickerMax = useMemo(
    () => toPickerString(parseDateTimeString(max ?? "")),
    [max]
  );

  const openPicker = useCallback(() => {
    if (disabled) {
      return;
    }
    const input = pickerRef.current;
    if (input == null) {
      return;
    }
    try {
      if (typeof input.showPicker === "function") {
        input.showPicker();
      } else {
        input.click();
      }
    } catch {
      input.click();
    }
  }, [disabled]);

  const onChangeText = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      onChange(event.target.value);
    },
    [onChange]
  );

  const onPickDateTime = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      // input[type=datetime-local] yields `yyyy-MM-ddTHH:mm` (or "").
      const raw = event.target.value;
      if (raw === "") {
        onChange("");
        return;
      }
      const picked = DateTime.fromFormat(
        raw,
        DATETIME_PICKER_FORMAT
      ).toJSDate();
      onChange(
        isNaN(picked.getTime())
          ? ""
          : DateTime.fromJSDate(picked).toFormat(DATETIME_DISPLAY_FORMAT)
      );
    },
    [onChange]
  );

  const pickerAriaLabel =
    ariaLabel ?? (typeof label === "string" ? label : placeholder);

  return (
    <TextField
      size={size}
      labelSize={labelSize}
      type="text"
      label={label}
      value={value}
      placeholder={placeholder}
      onChange={onChangeText}
      onClick={openPicker}
      disabled={disabled}
      inputClassName={styles.root}
      suffixPlain={true}
      suffix={
        <span className={styles.pickerWrap} onClick={openPicker}>
          <CalendarIcon
            className={styles.pickerIcon}
            width="1rem"
            height="1rem"
            aria-hidden={true}
          />
          <input
            ref={pickerRef}
            type="datetime-local"
            className={styles.nativePicker}
            value={pickerValue}
            min={pickerMin !== "" ? pickerMin : undefined}
            max={pickerMax !== "" ? pickerMax : undefined}
            onChange={onPickDateTime}
            disabled={disabled}
            tabIndex={-1}
            aria-label={pickerAriaLabel}
          />
        </span>
      }
    />
  );
}

export interface DateFieldDateTimeProps
  extends Omit<DateTimeFieldProps, "value" | "onChange" | "min" | "max"> {
  value: Date | null | undefined;
  onChange: (date: Date | null) => void;
  min?: Date;
  max?: Date;
}

/**
 * DateTimeField variant that speaks in `Date | null` instead of
 * `yyyy-MM-dd HH:mm` strings.
 */
export function DateFieldDateTime(
  props: DateFieldDateTimeProps
): React.ReactElement {
  const { value, onChange, min, max, ...rest } = props;

  const stringValue = useMemo(() => toDateTimeString(value), [value]);
  const minString = useMemo(() => toDateTimeString(min), [min]);
  const maxString = useMemo(() => toDateTimeString(max), [max]);

  const onChangeString = useCallback(
    (next: string) => {
      onChange(parseDateTimeString(next));
    },
    [onChange]
  );

  return (
    <DateTimeField
      {...rest}
      value={stringValue}
      onChange={onChangeString}
      min={minString !== "" ? minString : undefined}
      max={maxString !== "" ? maxString : undefined}
    />
  );
}

export { toDateString, parseDateString, toDateTimeString, parseDateTimeString };
