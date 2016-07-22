(function() {

  window.app = {};

  var labeling = null;
  var clearButton = null;
  var drawing = null;
  var classifier = null;

  function initialize() {
    window.app.loadClassifier('classifiers/forest', function(err, c) {
      if (err !== null) {
        showLoadError(err);
        return;
      }
      classifier = c;
      c.onLoad = initializeUI;
      c.onError = showLoadError;
    });
  }

  function initializeUI() {
    document.body.className = '';

    clearButton = document.getElementById('clear-button');
    clearButton.addEventListener('click', clear);

    labeling = document.getElementById('labeling');
    classifier.onClassify = function(classification) {
      labeling.className = '';
      labeling.textContent = classification;
    };

    drawing = new window.app.Drawing();
    drawing.onChange = function() {
      classifier.classify(drawing.mnistIntensities());
    };
  }

  function clear() {
    drawing.reset();
    classifier.cancel();
    labeling.className = 'hidden';
  }

  function showLoadError(err) {
    var l = document.getElementById('loading');
    l.textContent = 'Load failed: ' + err;
  }

  window.addEventListener('load', initialize);

})();
