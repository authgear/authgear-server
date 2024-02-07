import { Controller } from "@hotwired/stimulus";
import { autoUpdate, computePosition, flip } from "@floating-ui/dom";

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
    initialValue: String,
  };

  declare readonly inputTarget: HTMLInputElement;
  declare readonly triggerTarget: HTMLButtonElement;
  declare readonly dropdownTarget: HTMLElement;
  declare readonly searchTarget: HTMLInputElement;
  declare readonly clearSearchTarget: HTMLElement;
  declare readonly optionsTarget: HTMLElement;
  declare readonly searchTemplateTarget?: HTMLTemplateElement;
  declare readonly itemTemplateTarget: HTMLTemplateElement;
  declare readonly emptyTemplateTarget?: HTMLTemplateElement;

  declare readonly optionsValue: SearchSelectOption[];
  declare readonly initialValueValue: string;

  private isInitialized: boolean = false;

  get filteredOptions() {
    return this.optionsValue.filter((option) => {
      return `${option.label} ${option.value} ${option.prefix} ${option.searchLabel}`
        .toLocaleLowerCase()
        .includes(this.keyword.toLocaleLowerCase());
    });
  }

  get isOpen() {
    return !this.dropdownTarget.classList.contains("hidden");
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
    if (this.inputTarget.value === "") {
      this.inputTarget.value = this.initialValueValue;
    }
    this._computePositionCleanup = autoUpdate(
      this.triggerTarget,
      this.dropdownTarget,
      this._updateDropdownPosition
    );

    this.dropdownTarget.classList.add("hidden");
    this.renderTrigger();
    this.renderSearch();
    this.renderItems();

    this.triggerTarget.addEventListener("keydown", this.handleKeyDown);
    this.dropdownTarget.addEventListener("keydown", this.handleKeyDown);
    document.addEventListener("click", this.handleClickOutside);

    this.isInitialized = true;
  }

  disconnect(): void {
    this._computePositionCleanup();

    this.triggerTarget.removeEventListener("keydown", this.handleKeyDown);
    this.dropdownTarget.removeEventListener("keydown", this.handleKeyDown);
    document.removeEventListener("click", this.handleClickOutside);
  }

  optionsValueChanged() {
    if (!this.isInitialized) {
      return;
    }
    this.renderTrigger();
    this.renderItems();
  }

  open() {
    if (!this.dropdownTarget.classList.contains("hidden")) return;

    this.dropdownTarget.classList.remove("hidden");
    this.triggerTarget.setAttribute("aria-expanded", "true");

    this.clearSearch();
    this.resetScroll();

    this.searchTarget?.focus();
  }

  close() {
    if (this.dropdownTarget.classList.contains("hidden")) return;

    this.dropdownTarget.classList.add("hidden");
    this.triggerTarget.setAttribute("aria-expanded", "false");
    this.triggerTarget.focus();
  }

  toggle() {
    const willExpand = this.dropdownTarget.classList.contains("hidden");
    if (willExpand) {
      this.open();
    } else {
      this.close();
    }
  }

  search(event: InputEvent) {
    this.renderItems();
    this.resetScroll();
  }

  clearSearch() {
    this.searchTarget!.value = "";
    this.renderItems();
  }

  resetScroll() {
    const item = this.optionsTarget.querySelector<HTMLLIElement>(
      `[data-value="${this.focusedValue ?? this.value}"]`
    );
    if (item) {
      item.scrollIntoView({ block: "center" });
    }
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
      placement: "bottom-start",
      middleware: [
        flip({
          fallbackPlacements: ["top-start"],
          fallbackStrategy: "initialPlacement",
        }),
      ],
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
    if (item == null) {
      console.warn("Trying to select an option which does not exist");
      return;
    }
    this._updateAriaSelected(item);
    this.inputTarget.value = value ?? "";
    this.inputTarget.dispatchEvent(new Event("input", { bubbles: true }));

    this.renderTrigger();
    this.close();
  }

  _updateAriaSelected(item: HTMLLIElement) {
    this.optionsTarget.querySelectorAll('[role="option"]').forEach((option) => {
      option.setAttribute("aria-selected", "false");
    });
    item.setAttribute("aria-selected", "true");
    this._scrollIntoNearestView(item);
  }

  // Default `scrollIntoView({ block: "nearest" })` does not keep padding
  // into account, which makes the selected item stick to the top/bottom of the
  // dropdown.
  _scrollIntoNearestView(item: HTMLLIElement) {
    const container = this.optionsTarget;
    const containerPadding = parseFloat(getComputedStyle(container).paddingTop);
    const padding = parseFloat(getComputedStyle(item).paddingTop);
    const itemPosition = item.offsetTop - this.searchTarget.offsetHeight;

    let scrollPosition: number | undefined;

    switch (true) {
      case container.firstElementChild === item:
        scrollPosition = 0;
        break;
      case container.lastElementChild === item:
        scrollPosition = container.scrollHeight;
        break;
      case itemPosition < container.scrollTop + padding - containerPadding:
        scrollPosition = itemPosition - padding;
        break;
      case itemPosition + item.offsetHeight + padding >
        container.scrollTop + container.offsetHeight - containerPadding:
        scrollPosition =
          itemPosition + item.offsetHeight + padding - container.offsetHeight;
        break;
    }

    if (scrollPosition !== undefined) {
      requestAnimationFrame(() => {
        container.scrollTo({ top: scrollPosition });
      });
    }
  }

  renderTrigger() {
    const option =
      this.optionsValue.find((option) => option.value === this.value) ??
      this.optionsValue[0];
    this.triggerTarget.innerHTML = option?.triggerLabel ?? option?.label ?? "";
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
    if (!this.isOpen) {
      return;
    }

    const container = this.optionsTarget;
    const template = this.itemTemplateTarget.content;

    const fragment = document.createDocumentFragment();

    this.filteredOptions.forEach((item, index) => {
      const clone = document.importNode(template, true);
      const option = clone.querySelector("li");
      const selected = this.keyword ? index === 0 : item.value === this.value;
      const prefixEl = option!.querySelector<HTMLElement>(
        '[data-label="prefix"]'
      );
      const labelEl = option!.querySelector<HTMLElement>(
        '[data-label="content"]'
      );
      if (prefixEl) {
        prefixEl.style.pointerEvents = "none";
        prefixEl.innerHTML = item.prefix ?? "";
      }
      if (labelEl) {
        labelEl.style.pointerEvents = "none";
        labelEl.innerHTML = item.label;
      }
      if (!prefixEl && !labelEl) {
        option!.innerHTML = item.label;
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

  public select(value: string | undefined) {
    this._selectValue(value);
  }
}
