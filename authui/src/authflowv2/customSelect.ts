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
    "dropdownContainer",
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
  declare readonly dropdownContainerTarget: HTMLElement;
  declare readonly searchTarget: HTMLInputElement;
  declare readonly clearSearchTarget: HTMLElement;
  declare readonly optionsTarget: HTMLElement;
  declare readonly searchTemplateTarget?: HTMLTemplateElement;
  declare readonly itemTemplateTarget: HTMLTemplateElement;
  declare readonly emptyTemplateTarget?: HTMLTemplateElement;

  declare readonly hasSearchTarget: boolean;

  declare readonly optionsValue: SearchSelectOption[];
  declare readonly initialValueValue: string;

  private _highlightIndex: number = 0;
  private set highlightIndex(value: number) {
    this._highlightIndex = Math.max(
      0,
      Math.min(value, this.filteredOptions.length - 1)
    );
  }
  private get highlightIndex(): number {
    return this._highlightIndex;
  }

  private isInitialized: boolean = false;

  get filteredOptions(): SearchSelectOption[] {
    if (!this.keyword) {
      return this.optionsValue;
    }
    return this.optionsValue.filter((option) => {
      return `${option.label} ${option.value} ${option.searchLabel}`
        .toLocaleLowerCase()
        .includes(this.keyword.toLocaleLowerCase());
    });
  }

  get isOpen(): boolean {
    return !this.dropdownContainerTarget.classList.contains("hidden");
  }

  get value() {
    return this.inputTarget.value;
  }

  get keyword() {
    if (!this.hasSearchTarget) {
      return "";
    }

    return this.searchTarget.value;
  }

  get highlightedValue() {
    return (
      this.filteredOptions[this.highlightIndex] as
        | undefined
        | SearchSelectOption
    )?.value;
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

    this.dropdownContainerTarget.classList.add("hidden");
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
    if (!this.dropdownContainerTarget.classList.contains("hidden")) return;

    this.dropdownContainerTarget.classList.remove("hidden");
    this.triggerTarget.setAttribute("aria-expanded", "true");

    this.resetHightlightIndex();
    this.clearSearch();
    this.resetScroll();

    if (this.hasSearchTarget) {
      this.searchTarget.focus();
    }

    this.dispatch("open");
  }

  close() {
    if (this.dropdownContainerTarget.classList.contains("hidden")) return;

    this.dropdownContainerTarget.classList.add("hidden");
    this.triggerTarget.setAttribute("aria-expanded", "false");
    this.triggerTarget.focus();

    this.dispatch("close");
  }

  private resetHightlightIndex() {
    if (this.keyword) {
      this.highlightIndex = 0;
    } else {
      this.highlightIndex = this.optionsValue.findIndex(
        (o) => o.value === this.value
      );
    }
  }

  toggle() {
    const willExpand =
      this.dropdownContainerTarget.classList.contains("hidden");
    if (willExpand) {
      this.open();
    } else {
      this.close();
    }
  }

  search() {
    this.resetHightlightIndex();
    this.renderItems();
    this.resetScroll();
  }

  clearSearch() {
    if (!this.hasSearchTarget) return;

    this.searchTarget.value = "";
    this.renderItems();
  }

  resetScroll() {
    const item = this.optionsTarget.querySelector<HTMLLIElement>(
      `[data-value="${this.highlightedValue ?? this.value}"]`
    );
    if (item) {
      item.scrollIntoView({ block: "center" });
    }
  }

  navigate(indexFn: (index: number) => number) {
    this.highlightIndex = indexFn(this.highlightIndex);
    const item = this.optionsTarget.querySelector<HTMLLIElement>(
      `[data-value="${this.highlightedValue}"]`
    );

    if (!item) return;
    this._updateAriaSelected(item);
  }

  handleSelect(event: MouseEvent) {
    const item = event.target as HTMLLIElement | undefined;
    if (!item) return;

    const value = item.dataset.value;
    this._selectValue(value);
  }

  // eslint-disable-next-line complexity
  handleKeyDown = (event: KeyboardEvent) => {
    let preventAndStop = true;
    switch (event.key) {
      case "ArrowDown":
        this.navigate((idx) => idx + 1);
        break;
      case "ArrowUp":
        this.navigate((idx) => idx - 1);
        break;
      case "PageDown":
        this.navigate((idx) => idx + 10);
        break;
      case "PageUp":
        this.navigate((idx) => idx - 10);
        break;
      case "Home":
        this.navigate((_idx) => 0);
        break;
      case "End":
        this.navigate((_idx) => this.filteredOptions.length - 1);
        break;
      case "Enter":
        this._selectValue(this.highlightedValue ?? this.value);
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
    void computePosition(this.triggerTarget, this.dropdownTarget, {
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
    const option = this.optionsValue.find((o) => o.value === value);
    if (option == null) {
      console.warn("Trying to select an option which does not exist");
      return;
    }
    const optionEl = this.optionsTarget.querySelector<HTMLLIElement>(
      `[data-value="${option.value}"]`
    );
    if (optionEl != null) {
      this._updateAriaSelected(optionEl);
    }
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
    const itemPosition =
      item.offsetTop -
      (this.hasSearchTarget ? this.searchTarget.offsetHeight : 0);

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
      (this.optionsValue[0] as SearchSelectOption | undefined);

    if (!option) {
      return;
    }

    this.triggerTarget.innerHTML = option.triggerLabel ?? option.label;
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

    const visibleOptions = this.filteredOptions;

    const visibleOptionsIndexByValue = visibleOptions.reduce<
      Map<string, number>
    >((result, option, idx, _arr): Map<string, number> => {
      result.set(option.value, idx);
      return result;
    }, new Map());

    this.optionsValue.forEach((item, index) => {
      const clone = document.importNode(template, true);
      const option = clone.querySelector("li");
      const isVisible = visibleOptionsIndexByValue.get(item.value) != null;
      const selected =
        visibleOptionsIndexByValue.get(item.value) === this.highlightIndex;
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
      if (!isVisible) {
        option!.classList.add("hidden");
      }
      fragment.appendChild(clone);
    });

    if (visibleOptions.length === 0 && this.emptyTemplateTarget) {
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
