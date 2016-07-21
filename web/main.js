(function() {

  window.app = {};

  function initialize() {
    var d = new window.app.Drawing();
    d.onChange = function() {
      console.log('drew bitmap.');
    };
  }

  window.addEventListener('load', initialize);

})();
