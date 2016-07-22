(function() {

  window.app = {};

  var labeling = null;
  var clearButton = null;
  var drawing = null;
  var classifier = null;
  var drawingEmpty = true;

  function initialize() {
    window.app.loadClassifier('classifiers/forest', classifierLoaded);

    clearButton = document.getElementById('clear-button');
    clearButton.addEventListener('click', clear);

    labeling = document.getElementById('labeling');

    drawing = new window.app.Drawing();
    drawing.onChange = function() {
      drawingEmpty = false;
      classifier.classify(drawing.mnistIntensities());
    };
  }

  function clear() {
    drawingEmpty = true;
    drawing.reset();
    classifier.cancel();
    labeling.className = 'hidden';
  }

  function classifierLoaded(err, c) {
    if (err !== null) {
      alert('Failed to load classifier: ' + err);
    } else {
      classifier = c;
      c.onClassify = function(classification) {
        labeling.className = '';
        labeling.textContent = classification;
      };
      if (!drawingEmpty) {
        c.classify(drawing.mnistIntensities());
      }
    }
  }

  window.addEventListener('load', initialize);

})();
