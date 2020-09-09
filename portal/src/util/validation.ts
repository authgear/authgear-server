export interface ValidationRule<Data, ErrorState> {
  inputKey: keyof Data;
  errorKey: keyof ErrorState;
  condition: (input: any) => boolean;
  errorMessageId: string;
}

export type ErrorResult<ErrorState> = Partial<
  { [key in keyof ErrorState]: string }
>;

export function validateInput<Data, ErrorState>(
  data: Data,
  validationRules: ValidationRule<Data, ErrorState>[]
): {
  errorResult: ErrorResult<ErrorState>;
  isValid: boolean;
} {
  let isValid = true;
  const errorResult: ErrorResult<ErrorState> = {};
  validationRules.forEach(
    (validationRule: ValidationRule<Data, ErrorState>) => {
      if (!validationRule.condition(data[validationRule.inputKey])) {
        errorResult[validationRule.errorKey] = validationRule.errorMessageId;
        isValid = false;
      }
    }
  );
  return { errorResult, isValid };
}

export function isValidEmail(email: string): boolean {
  const emailRegExp = /^.+@.+$/;
  return emailRegExp.test(email);
}

export function isValidEmailDomain(domain: string): boolean {
  return domain !== "";
}

export function isErrorResult<ErrorState>(
  result: any
): result is ErrorResult<ErrorState> {
  if (!(result instanceof Object)) {
    return false;
  }
  return Object.keys(result).length > 0;
}

export function renderErrorMessage(
  renderToString: (messageId: string) => string,
  errorMessageId?: string
): string | undefined {
  if (errorMessageId == null) {
    return undefined;
  }
  return renderToString(errorMessageId);
}
