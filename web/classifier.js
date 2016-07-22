(function() {

  var WEBWORKER_FILE = 'webworker/webworker.js';
  var XHR_DONE = 4;
  var HTTP_OK = 200;

  function Classifier(classifierData) {
    this.onClassify = null;
    this.onLoad = null;
    this.onError = null;

    this._queued = null;
    this._isRunning = false;
    this._canceled = false;

    this._worker = new Worker(WEBWORKER_FILE);
    this._worker.onmessage = function(e) {
      if ('undefined' !== typeof e.data.init) {
        if (e.data.init === null) {
          this.onLoad();
        } else {
          this.onError(e.data.init);
        }
        return;
      }
      this._isRunning = false;
      if (this._queued !== null) {
        var q = this._queued;
        this._queued = null;
        this.classify(q);
      } else if (!this._canceled && 'function' === typeof this.onClassify) {
        this.onClassify(e.data.classification);
      }
    }.bind(this);
    this._worker.postMessage(['init', classifierData]);
  }

  Classifier.prototype.classify = function(sample) {
    this._canceled = false;
    if (this._isRunning) {
      this._queued = sample;
    } else {
      this._isRunning = true;
      this._worker.postMessage(['classify', sample]);
    }
  };

  Classifier.prototype.cancel = function() {
    this._queued = null;
    this._canceled = true;
  };

  function loadClassifier(path, callback) {
    fetchClassifier(path, function(err, data) {
      if (err !== null) {
        callback(err, null);
      } else {
        callback(null, new Classifier(data));
      }
    });
  }

  function fetchClassifier(path, callback) {
    var xhr = new XMLHttpRequest();
    xhr.responseType = "arraybuffer";
    xhr.open('GET', path);
    xhr.send(null);

    xhr.onreadystatechange = function () {
      if (xhr.readyState === XHR_DONE) {
        if (xhr.status === HTTP_OK) {
          callback(null, new Uint8Array(xhr.response));
        } else {
          callback('status '+xhr.status, null);
        }
      }
    };
  }

  window.app.loadClassifier = loadClassifier;

})();
