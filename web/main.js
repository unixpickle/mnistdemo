(function() {

  window.app = {};

  var clearButton = null;
  var drawing = null;

  function initialize() {
    clearButton = document.getElementById('clear-button');
    clearButton.addEventListener('click', clear);

    drawing = new window.app.Drawing();
    drawing.onChange = function() {
      clearButton.className = '';
    };
  }

  function clear() {
    drawing.reset();
    clearButton.className = 'hidden';
  }

  window.addEventListener('load', initialize);

})();
