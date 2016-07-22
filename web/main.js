(function() {

  window.app = {};

  function initialize() {
    var d = new window.app.Drawing();
    d.onChange = function() {
      // Code is usedful to see that the intensities are valid.
      var bmp = d.mnistIntensities();
      var canvas = document.createElement('canvas');
      canvas.width = 28;
      canvas.height = 28;
      var ctx = canvas.getContext('2d');
      ctx.fillStyle = 'black';
      var idx = 0;
      for (var y = 0; y < 28; ++y) {
        for (var x = 0; x < 28; ++x) {
          ctx.globalAlpha = 1 - bmp[idx++];
          ctx.fillRect(x, y, 1, 1);
        }
      }
      canvas.style.border = '1px solid black';
      document.body.appendChild(canvas);
    };
  }

  window.addEventListener('load', initialize);

})();
