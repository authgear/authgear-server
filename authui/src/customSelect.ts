import { Controller } from "@hotwired/stimulus";
import { autoPlacement, computePosition, flip, shift } from "@floating-ui/dom";

export interface SearchSelectOption {
  triggerLabel?: string;
  label: string;
  value: string;
}

export class CustomSelectController extends Controller {
  static targets = [
    "trigger",
    "dropdown",
    "options",
    "searchTemplate",
    "itemTemplate",
  ];
  static values = {
    options: Array,
    default: String,
  };

  declare readonly triggerTarget: HTMLButtonElement;
  declare readonly dropdownTarget: HTMLElement;
  declare readonly optionsTarget: HTMLElement;
  declare readonly searchTemplateTarget?: HTMLTemplateElement;
  declare readonly itemTemplateTarget: HTMLTemplateElement;

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
  selectedItemIndex: number = -1;

  connect(): void {
    this.dropdownTarget.classList.add("hidden");
    computePosition(this.triggerTarget, this.dropdownTarget, {
      middleware: [flip(), shift(), autoPlacement({ alignment: "start" })],
    }).then(({ x, y }) => {
      Object.assign(this.dropdownTarget.style, {
        left: `${x}px`,
        top: `${y}px`,
      });
    });

    this.renderSearch();
    this.renderItems();
    this.triggerTarget.textContent =
      this.optionsValue[0]?.triggerLabel ?? this.optionsValue[0].label ?? "";

    this.triggerTarget.addEventListener("keydown", this.handleKeyDown);
    this.dropdownTarget.addEventListener("keydown", this.handleKeyDown);
    document.addEventListener("click", this.handleClickOutside);
  }

  disconnect(): void {
    this.triggerTarget.addEventListener("keydown", this.handleKeyDown);
    this.dropdownTarget.removeEventListener("keydown", this.handleKeyDown);
    document.removeEventListener("click", this.handleClickOutside);
  }

  optionsValueChanged() {}

  close() {
    this.dropdownTarget.classList.add("hidden");
    this.triggerTarget.setAttribute("aria-expanded", "false");
  }

  toggle() {
    const expanded = this.dropdownTarget.classList.toggle("hidden");
    this.triggerTarget.setAttribute("aria-expanded", expanded.toString());
    if (expanded) {
      const items = this.optionsTarget.querySelectorAll<any>('[role="option"]');
      const selectedItem = items[this.selectedItemIndex] ?? items[0];
      selectedItem?.focus();
    } else {
      this.triggerTarget.focus();
    }
  }

  clear() {
    const searchInput =
      this.searchTemplateTarget?.querySelector<HTMLInputElement>("input");
    console.log("clear", searchInput, this.searchTemplateTarget);
    if (!searchInput) return;

    searchInput.value = "";
    this.keyword = "";
    this.renderItems();
  }

  selectItem(event: any) {
    const itemValue = event.target?.getAttribute("data-value");
    const option = this.optionsValue.find(
      (option) => option.value === itemValue
    );
    this.triggerTarget.textContent =
      option?.triggerLabel ?? option?.label ?? "";
    this.toggle();
    this.updateAriaSelected(event.target);
  }

  selectItemByIndex() {
    const item = this.optionsTarget.querySelector<HTMLLIElement>(
      `[data-index="${this.selectedItemIndex}"]`
    );
    if (!item) return;
    this.selectItem({ target: item });
  }

  updateAriaSelected(selectedItem: HTMLLIElement) {
    this.optionsTarget.querySelectorAll('[role="option"]').forEach((option) => {
      option.setAttribute("aria-selected", "false");
    });
    selectedItem.setAttribute("aria-selected", "true");
  }

  navigate(step: number) {
    const items = this.optionsTarget.querySelectorAll<any>('[role="option"]');
    if (!items.length) return;

    this.selectedItemIndex =
      (this.selectedItemIndex + step + items.length) % items.length;
    const index = this.selectedItemIndex;

    requestAnimationFrame(() => {
      if (index !== this.selectedItemIndex) {
        return;
      }

      const selectedItem = items[this.selectedItemIndex];
      selectedItem?.focus();
      this.updateAriaSelected(selectedItem);
    });
  }

  handleKeyDown = (event: KeyboardEvent) => {
    switch (event.key) {
      case "ArrowDown":
        this.navigate(1);
        event.preventDefault();
        event.stopPropagation();
        break;
      case "ArrowUp":
        this.navigate(-1);
        event.preventDefault();
        event.stopPropagation();
        break;
      case "Enter":
        this.selectItemByIndex();
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

  renderSearch() {
    const container = this.dropdownTarget;
    if (!this.searchTemplateTarget) {
      return;
    }

    const template = this.searchTemplateTarget.content;

    const clone = document.importNode(template, true);
    const search = clone.querySelector("input");
    search?.addEventListener("input", (event) => {
      this.keyword = (event.target as HTMLInputElement).value;
      this.renderItems();
    });
    const clear = clone.querySelector("button");
    clear?.addEventListener("click", () => {
      this.clear();
    });
    container.prepend(clone);
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

    this.optionsTarget.innerHTML = "";
    container.appendChild(fragment);
  }
}
