(function() {
  document.body.addEventListener('htmx:pushedIntoHistory', () => {
    localStorage.removeItem('htmx-history-cache')
  })
})()
