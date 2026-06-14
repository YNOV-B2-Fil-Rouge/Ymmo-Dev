// Shared accessibility helpers.

// Turn toggle buttons + panels into a WAI-ARIA tab interface (tablist/tab/tabpanel).
// `tabs`: buttons with data-tab="<name>"; `panels`: { name -> #panel-<name> element }.
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
