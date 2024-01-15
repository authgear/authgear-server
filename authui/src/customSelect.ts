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
  label: string;
  value: string;
}

export class CustomSelectController extends Controller {
  static targets = [
    "input",
    "trigger",
    "dropdown",
    "search",
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
  declare readonly dropdownTarget: HTMLElement;
  declare readonly searchTarget?: HTMLInputElement;
  declare readonly optionsTarget: HTMLElement;
  declare readonly searchTemplateTarget?: HTMLTemplateElement;
  declare readonly itemTemplateTarget: HTMLTemplateElement;
  declare readonly emptyTemplateTarget?: HTMLTemplateElement;

  declare readonly optionsValue: SearchSelectOption[];

  get filteredOptions() {
    return this.optionsValue.filter((option) => {
      return (
        option.label.toLowerCase().includes(this.keyword.toLowerCase()) ||
        option.value?.toLowerCase().includes(this.keyword.toLowerCase())
      );
    });
  }

  keyword = "";
  value?: string;

  _computePositionCleanup = () => {};

  connect(): void {
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
  }

  disconnect(): void {
    this._computePositionCleanup();

    this.triggerTarget.addEventListener("keydown", this.handleKeyDown);
    this.dropdownTarget.removeEventListener("keydown", this.handleKeyDown);
    document.removeEventListener("click", this.handleClickOutside);
  }

  open() {
    if (!this.dropdownTarget.classList.contains("hidden")) return;

    this.dropdownTarget.classList.remove("hidden");
    this.triggerTarget.setAttribute("aria-expanded", "true");

    const item = this.optionsTarget.querySelector<HTMLLIElement>(
      `[data-value="${this.value}"]`
    );
    item?.focus();
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
    const searchInput = event.target as HTMLInputElement;
    this.keyword = searchInput.value;
    this.renderItems();
  }

  clear() {
    const searchInput =
      this.searchTarget?.querySelector<HTMLInputElement>("input");
    if (!searchInput) return;

    searchInput.value = "";
    this.keyword = "";
    this.renderItems();
  }

  navigate(step: number) {
    const currentIndex = this.filteredOptions.findIndex(
      (option) => option.value === this.value
    );
    const newIndex =
      (currentIndex + step + this.filteredOptions.length) %
      this.filteredOptions.length;
    const newValue = this.filteredOptions[newIndex]?.value;

    requestAnimationFrame(() => {
      if (newValue !== this.value) {
        return;
      }

      const selectedItem = this.optionsTarget.querySelector<HTMLLIElement>(
        `[data-value="${newValue}"]`
      );
      if (!selectedItem) return;

      selectedItem.focus();
      this._updateAriaSelected(selectedItem);
    });
  }

  handleSelect(event: MouseEvent) {
    const item = event.target as HTMLLIElement;
    if (!item) return;

    const value = item.dataset.value;
    this._selectValue(value);
  }

  handleKeyDown = (event: KeyboardEvent) => {
    switch (event.key) {
      case "ArrowDown":
        this.navigate(1);
        event.stopPropagation();
        break;
      case "ArrowUp":
        this.navigate(-1);
        event.stopPropagation();
        break;
      case "Enter":
        this._selectValue();
        break;
      case "Escape":
        this.close();
        break;
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

  _selectValue(value: string | undefined = this.value) {
    const item = this.optionsTarget.querySelector<HTMLLIElement>(
      `[data-value="${value}"]`
    );
    if (!item) return;
    this.value = value;
    this._updateAriaSelected(item);
    this.renderTrigger();
    this.close();
    this.inputTarget.value = value ?? "";
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
    const container = this.dropdownTarget;
    if (!this.searchTemplateTarget) {
      return;
    }

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
      option!.textContent = item.label;
      option!.dataset.index = index.toString();
      option!.setAttribute("data-value", item.value);
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
