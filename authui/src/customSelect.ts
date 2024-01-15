import { Controller } from "@hotwired/stimulus";
import { autoPlacement, computePosition, flip, shift } from "@floating-ui/dom";

export interface SearchSelectOption {
  triggerLabel?: string;
  label: string;
  value: string;
}

export class CustomSelectController extends Controller {
  static targets = ["trigger", "dropdown", "options", "itemTemplate"];
  static values = {
    options: Array,
    default: String,
  };

  declare readonly triggerTarget: HTMLButtonElement;
  declare readonly dropdownTarget: HTMLElement;
  declare readonly optionsTarget: HTMLElement;
  declare readonly itemTemplateTarget: HTMLTemplateElement;

  declare readonly optionsValue: SearchSelectOption[];

  _selectedItemIndex: number = -1;

  connect(): void {
    this.dropdownTarget.classList.add("hidden");
    computePosition(this.triggerTarget, this.dropdownTarget, {
      placement: "top",
      middleware: [flip(), shift(), autoPlacement({ alignment: "start" })],
    }).then(({ x, y }) => {
      Object.assign(this.dropdownTarget.style, {
        left: `${x}px`,
        top: `${y}px`,
      });
    });

    this.triggerTarget.addEventListener("keydown", this.handleKeyDown);
    this.dropdownTarget.addEventListener("keydown", this.handleKeyDown);
    document.addEventListener("click", this.handleClickOutside);
  }

  disconnect(): void {
    this.triggerTarget.addEventListener("keydown", this.handleKeyDown);
    this.dropdownTarget.removeEventListener("keydown", this.handleKeyDown);
    document.removeEventListener("click", this.handleClickOutside);
  }

  optionsValueChanged() {
    this.renderItems();
    this.triggerTarget.textContent = this.optionsValue[0]?.triggerLabel ?? this.optionsValue[0].label ?? "";
  }

  close() {
    this.dropdownTarget.classList.add("hidden");
    this.triggerTarget.setAttribute("aria-expanded", "false");
  }

  toggle() {
    const expanded = this.dropdownTarget.classList.toggle("hidden");
    this.triggerTarget.setAttribute("aria-expanded", expanded.toString());
    if (expanded) {
      const items = this.optionsTarget.querySelectorAll<any>('[role="option"]');
      const selectedItem = items[this._selectedItemIndex] ?? items[0];
      selectedItem?.focus();
    } else {
      this.triggerTarget.focus();
    }
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
      `[data-index="${this._selectedItemIndex}"]`
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

    this._selectedItemIndex =
      (this._selectedItemIndex + step + items.length) % items.length;
    const index = this._selectedItemIndex;

    requestAnimationFrame(() => {
      if (index !== this._selectedItemIndex) {
        return;
      }

      const selectedItem = items[this._selectedItemIndex];
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

  renderItems() {
    const container = this.optionsTarget;
    const template = this.itemTemplateTarget.content;

    const fragment = document.createDocumentFragment();

    this.optionsValue.forEach((item, index) => {
      const clone = document.importNode(template, true);
      const option = clone.querySelector("li");
      option!.textContent = item.label;
      option!.dataset.index = index.toString();
      option!.setAttribute("data-value", item.value);
      fragment.appendChild(clone);
    });

    container.appendChild(fragment);
  }
}
