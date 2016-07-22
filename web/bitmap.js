(function() {

  function Bitmap(canvas) {
    var ctx = canvas.getContext('2d');
    this._data = ctx.getImageData(0, 0, canvas.width, canvas.height);
  }

  Bitmap.prototype.usedBounds = function() {
    var minX = this._data.width;
    var minY = this._data.height;
    var maxX = 0;
    var maxY = 0;

    var idx = 3;
    for (var y = 0; y < this._data.height; ++y) {
      for (var x = 0; x < this._data.width; ++x) {
        var pixel = this._data.data[idx];
        if (pixel > 0) {
          minX = Math.min(minX, x);
          minY = Math.min(minY, y);
          maxX = Math.max(maxX, x);
          maxY = Math.max(maxY, y);
        }
        idx += 4;
      }
    }

    if (minX >= maxX || minY >= maxY) {
      return {x: 0, y: 0, width: 0, height: 0};
    }

    return {
      x: minX,
      y: minY,
      width: maxX-minX,
      height: maxY-minY
    };
  };

  Bitmap.prototype.centerOfMass = function() {
    var totalMass = this._totalMass();

    var xCenter = 0;
    for (var x = 0; x < this._data.width; ++x) {
      var columnMass = 0;
      for (var y = 0; y < this._data.height; ++y) {
        columnMass += this._massAtPoint(x, y);
      }
      xCenter += columnMass * x;
    }
    xCenter /= totalMass;

    var yCenter = 0;
    for (var y = 0; y < this._data.height; ++y) {
      var rowMass = 0;
      for (var x = 0; x < this._data.width; ++x) {
        rowMass += this._massAtPoint(x, y);
      }
      yCenter += rowMass * y;
    }
    yCenter /= totalMass;

    return {x: xCenter, y: yCenter};
  };

  Bitmap.prototype.alphaData = function() {
    var data = [];
    var idx = 3;
    while (idx < this._data.data.length) {
      data.push(this._data.data[idx]);
      idx += 4;
    }
    return data;
  };

  Bitmap.prototype._massAtPoint = function(x, y) {
    return this._data.data[3 + 4*(x+y*this._data.width)];
  };

  Bitmap.prototype._totalMass = function() {
    var totalMass = 0;
    var idx = 3;
    for (var y = 0; y < this._data.height; ++y) {
      for (var x = 0; x < this._data.width; ++x) {
        totalMass += this._data.data[idx];
        idx += 4;
      }
    }
    return totalMass;
  };

  window.app.Bitmap = Bitmap;

})();
