class ChoiceButton extends HTMLElement {
  constructor() {
    super()
    const choice = this.getAttribute("choice")
    const disable = this.getAttribute("disabled") ?? false
    this.innerHTML = `
      <form ws-send hx-trigger="click" ${disable ?? 'hx-disable'}>
        <div>
          <input type="hidden" name="choice" value="${choice}" />
          <img src="./images/icon-${choice}.svg" alt="${choice}" />
        </div>
      </form>
    `
  }
}

customElements.define("choice-button", ChoiceButton)
