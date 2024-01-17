import { Controller } from "@hotwired/stimulus";
import {
  autoPlacement,
  autoUpdate,
  computePosition,
  flip,
  shift,
} from "@floating-ui/dom";

export interface SearchSelectOption {
  triggerLabel?: string;
  prefix?: string;
  label: string;
  searchLabel?: string;
  value: string;
}

export class CustomSelectController extends Controller {
  static targets = [
    "input",
    "trigger",
    "dropdown",
    "search",
    "clearSearch",
    "options",
    "searchTemplate",
    "itemTemplate",
    "emptyTemplate",
  ];
  static values = {
    options: Array,
  };

  declare readonly inputTarget: HTMLInputElement;
  declare readonly triggerTarget: HTMLButtonElement;
  declare readonly dropdownTarget: HTMLDialogElement;
  declare readonly searchTarget: HTMLInputElement;
  declare readonly clearSearchTarget: HTMLElement;
  declare readonly optionsTarget: HTMLElement;
  declare readonly searchTemplateTarget?: HTMLTemplateElement;
  declare readonly itemTemplateTarget: HTMLTemplateElement;
  declare readonly emptyTemplateTarget?: HTMLTemplateElement;

  declare readonly optionsValue: SearchSelectOption[];

  get filteredOptions() {
    return this.optionsValue.filter((option) => {
      return `${option.label} ${option.value} ${option.prefix} ${option.searchLabel}`
        .toLocaleLowerCase()
        .includes(this.keyword.toLocaleLowerCase());
    });
  }

  get value() {
    return this.inputTarget.value;
  }

  get keyword() {
    return this.searchTarget.value ?? "";
  }

  get focusedValue() {
    return this.optionsTarget.querySelector<HTMLLIElement>(
      '[aria-selected="true"]'
    )?.dataset.value;
  }

  _computePositionCleanup = () => {};

  connect(): void {
    this._computePositionCleanup = autoUpdate(
      this.triggerTarget,
      this.dropdownTarget,
      this._updateDropdownPosition
    );

    this.renderTrigger();
    this.renderSearch();
    this.renderItems();

    this.triggerTarget.addEventListener("keydown", this.handleKeyDown);
    this.dropdownTarget.addEventListener("keydown", this.handleKeyDown);
    document.addEventListener("click", this.handleClickOutside);
  }

  disconnect(): void {
    this._computePositionCleanup();

    this.triggerTarget.addEventListener("keydown", this.handleKeyDown);
    this.dropdownTarget.removeEventListener("keydown", this.handleKeyDown);
    document.removeEventListener("click", this.handleClickOutside);
  }

  optionValuesChanged() {
    this.renderTrigger();
    this.renderItems();
  }

  open() {
    if (this.dropdownTarget.open) return;

    this.dropdownTarget.show();
    this.triggerTarget.setAttribute("aria-expanded", "true");

    this.clearSearch();

    if (!this.value) {
      this.searchTarget?.focus();
    }
  }

  close() {
    if (!this.dropdownTarget.open) return;

    this.dropdownTarget.close();
    this.triggerTarget.setAttribute("aria-expanded", "false");
    this.triggerTarget.focus();
  }

  toggle() {
    const willExpand = !this.dropdownTarget.open;
    if (willExpand) {
      this.open();
    } else {
      this.close();
    }
  }

  search(event: InputEvent) {
    if (this.keyword.length === 0) {
      this.clearSearchTarget.classList.add("hidden");
    } else {
      this.clearSearchTarget.classList.remove("hidden");
    }

    this.renderItems();
  }

  clearSearch() {
    this.searchTarget!.value = "";
    this.clearSearchTarget.classList.add("hidden");

    this.renderItems();
  }

  navigate(stepFn: (index: number) => number) {
    const currentIndex = this.filteredOptions.findIndex(
      (option) => option.value === (this.focusedValue ?? this.value)
    );
    const step = stepFn(currentIndex);
    const newIndex =
      (currentIndex + step + this.filteredOptions.length) %
      this.filteredOptions.length;
    const item = this.optionsTarget.querySelector<HTMLLIElement>(
      `[data-index="${newIndex}"]`
    );

    if (!item) return;
    this._updateAriaSelected(item);
    item.focus();
  }

  handleSelect(event: MouseEvent) {
    const item = event.target as HTMLLIElement;
    if (!item) return;

    const value = item.dataset.value;
    this._selectValue(value);
  }

  handleKeyDown = (event: KeyboardEvent) => {
    let preventAndStop = true;
    switch (event.key) {
      case "ArrowDown":
        this.navigate(() => 1);
        break;
      case "ArrowUp":
        this.navigate(() => -1);
        break;
      case "PageDown":
        this.navigate(() => 10);
        break;
      case "PageUp":
        this.navigate(() => -10);
        break;
      case "Home":
        this.navigate((idx) => -idx);
        break;
      case "End":
        this.navigate((idx) => this.filteredOptions.length - 1 - idx);
        break;
      case "Enter":
        this._selectValue(this.focusedValue ?? this.value);
        break;
      case "Escape":
        this.close();
        preventAndStop = false;
        break;
      default:
        preventAndStop = false;
    }

    if (preventAndStop) {
      event.preventDefault();
      event.stopPropagation();
    }
  };

  handleClickOutside = (event: MouseEvent) => {
    if (!this.element.contains(event.target as Node)) {
      this.close();
    }
  };

  _updateDropdownPosition = () => {
    computePosition(this.triggerTarget, this.dropdownTarget, {
      middleware: [flip(), shift(), autoPlacement({ alignment: "start" })],
    }).then(({ x, y }) => {
      Object.assign(this.dropdownTarget.style, {
        left: `${x}px`,
        top: `${y}px`,
      });
    });
  };

  _selectValue(value: string | undefined) {
    const item = this.optionsTarget.querySelector<HTMLLIElement>(
      `[data-value="${value}"]`
    );
    if (!item) return;
    this._updateAriaSelected(item);
    this.inputTarget.value = value ?? "";
    this.inputTarget.dispatchEvent(new Event("input", { bubbles: true }));

    this.renderTrigger();
    this.close();
  }

  _updateAriaSelected(selectedItem: HTMLLIElement) {
    this.optionsTarget.querySelectorAll('[role="option"]').forEach((option) => {
      option.setAttribute("aria-selected", "false");
    });
    selectedItem.setAttribute("aria-selected", "true");
  }

  renderTrigger() {
    const option =
      this.optionsValue.find((option) => option.value === this.value) ??
      this.optionsValue[0];
    this.triggerTarget.textContent =
      option?.triggerLabel ?? option?.label ?? "";
  }

  renderSearch() {
    if (!this.searchTemplateTarget) {
      return;
    }

    const container = this.dropdownTarget;
    const template = this.searchTemplateTarget.content;
    container.prepend(document.importNode(template, true));
  }

  renderItems() {
    const container = this.optionsTarget;
    const template = this.itemTemplateTarget.content;

    const fragment = document.createDocumentFragment();

    this.filteredOptions.forEach((item, index) => {
      const clone = document.importNode(template, true);
      const option = clone.querySelector("li");
      const selected = item.value === this.value;
      const prefixEl = option!.querySelector<HTMLElement>(
        '[data-label="prefix"]'
      );
      const labelEl = option!.querySelector<HTMLElement>(
        '[data-label="content"]'
      );
      if (prefixEl) {
        prefixEl.style.pointerEvents = "none";
        prefixEl.textContent = item.prefix ?? "";
      }
      if (labelEl) {
        labelEl.style.pointerEvents = "none";
        labelEl.textContent = item.label;
      }
      if (!prefixEl && !labelEl) {
        option!.textContent = item.label;
      }
      option!.dataset.index = index.toString();
      option!.setAttribute("data-value", item.value);
      option!.setAttribute("aria-selected", selected.toString());
      fragment.appendChild(clone);
    });

    if (this.filteredOptions.length === 0 && this.emptyTemplateTarget) {
      const emptyTemplate = this.emptyTemplateTarget.content;
      const clone = document.importNode(emptyTemplate, true);
      fragment.appendChild(clone);
    }

    this.optionsTarget.innerHTML = "";
    container.appendChild(fragment);
  }
}
