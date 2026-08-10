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

export { toDateString, parseDateString };
