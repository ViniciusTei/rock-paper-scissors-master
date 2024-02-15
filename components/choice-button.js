class ChoiceButton extends HTMLElement {
  constructor() {
    super()
    const choice = this.getAttribute("choice")
    this.innerHTML = `
      <form ws-send hx-trigger="click">
        <div>
          <input type="hidden" name="choice" value="${choice}" />
          <img src="./images/icon-${choice}.svg" alt="${choice}" />
        </div>
      </form>
    `
  }
}

customElements.define("choice-button", ChoiceButton)
