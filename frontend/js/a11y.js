// Shared accessibility helpers.

// enhanceTabsAria upgrades a set of toggle buttons + their panels into an
// accessible WAI-ARIA tab interface:
//   - the buttons are wrapped in a role="tablist" container,
//   - each button becomes role="tab" with an id + aria-controls,
//   - each panel becomes role="tabpanel" with aria-labelledby.
// The caller's showTab() keeps toggling aria-selected on the active tab.
//
// `tabs` is a NodeList/array of buttons carrying data-tab="<name>".
// `panels` is an object mapping <name> -> the panel element (id="panel-<name>").
export function enhanceTabsAria(tabs, panels) {
  const first = tabs[0];
  if (first && first.parentElement) {
    const list = document.createElement("div");
    list.setAttribute("role", "tablist");
    list.setAttribute("aria-orientation", "vertical");
    list.className = "space-y-3";
    first.parentElement.insertBefore(list, first);
    tabs.forEach((t) => {
      const name = t.dataset.tab;
      t.setAttribute("role", "tab");
      t.id = `tab-${name}`;
      t.setAttribute("aria-controls", `panel-${name}`);
      list.appendChild(t); // moves the button into the tablist (keeps listeners)
    });
  }
  Object.entries(panels).forEach(([name, el]) => {
    if (!el) return;
    el.setAttribute("role", "tabpanel");
    el.setAttribute("aria-labelledby", `tab-${name}`);
    el.setAttribute("tabindex", "0");
  });
}
