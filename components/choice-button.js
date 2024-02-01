class ChoiceButton extends HTMLElement {
  constructor() {
    super()
    const choice = this.getAttribute("choice")
    this.innerHTML = `
      <div 
        hx-get="/choose/${choice}" 
        hx-trigger="click" 
        hx-target="#main"
        hx-swap="outerHTML"
      >
        <img src="./images/icon-${choice}.svg" alt="${choice}" />
      </div>
    `
  }
}

customElements.define("choice-button", ChoiceButton)
