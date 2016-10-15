(function(){
  var audio_context;

  function __log(e, data) {
    console.log(e + " " + (data || ''));
  }

  function createRecorder(stream, handleMessage) {
    var input = audio_context.createMediaStreamSource(stream);

    var recorder = new Recorder(input, {
      serverUrl: "ws://54.223.187.43:8281/llcup/stream/upload",
      handleMessage: handleMessage
    });

    __log('Recorder initialised.');

    return recorder;
  }

  var Page = function() {
    var self = this;
    var inited = false;
    var recorder = null;

    var handleMessage = function(resp) {
      try {
        var respObj = JSON.parse(resp);
        self.overallScore(respObj.decoded);
        respObj.details.forEach(function(wordRate) {
          self.wordRates.push({
            word: wordRate.word,
            score: wordRate.confidence
          })
        });
      } catch (e) {
        self.hasError(true);
        self.errorResp(resp);
        self.errorInfo(e.message);
      }
    }

    this.inited = ko.observable(false);
    initAudioSetting(function(stream){
      recorder = createRecorder(stream, handleMessage);
      self.inited(true);
    });

    this.hasError = ko.observable(false);
    this.errorResp = ko.observable('');
    this.errorInfo = ko.observable('');
    this.wordRates = ko.observableArray([]);
    this.readingRefText = ko.observable(randomPick(Constants.PreparedTexts));
    this.recording = ko.observable(false);
    this.overallScore = ko.observable();
    this.recordButtonText = ko.computed(function() {
      return self.recording() ? "停止录音" : "开始录音";
    });
    this.toggleRecording = function() {
      self.hasError(false);
      self.wordRates.removeAll();
      self.recording(!self.recording());
    }

    //this.switchRefText = function() {
    //  self.readingRefText(randomPick(Constants.PreparedTexts));
    //}

    this.recording.subscribe(function(){
      if(self.recording()) {
/*
        algConfig = {
          type: 'readaloud',
          quality: -1,
          //reftext: self.readingRefText().replace(/[,.]/g, '')
          reference: self.readingRefText().toLowerCase().replace(/[^A-Za-z0-9']/g, ' ').trim()
        };
*/
        algConfig = {
          type: 'asr',
          quality: -1
        };
        console.log(algConfig);
        recorder.record({
          algConfig: algConfig
        });
      } else {
        recorder.stop();
        recorder.clear();
      }
    });
  }

  var initAudioSetting = function(startUserMediaCallback) {
    try {
      // webkit shim
      window.AudioContext = window.AudioContext || window.webkitAudioContext;
      navigator.getUserMedia = navigator.getUserMedia || navigator.webkitGetUserMedia;
      window.URL = window.URL || window.webkitURL;

      audio_context = new AudioContext;
      __log('Audio context set up.');
      __log('navigator.getUserMedia ' + (navigator.getUserMedia ? 'available.' : 'not present!'));
    } catch (e) {
      alert('No web audio support in this browser!');
    }

    navigator.getUserMedia({audio: true, video: false}, startUserMediaCallback, function(e) {
      __log('No live audio input: ' + e);
    });
  }

  window.onload = function init() {
    window.page = new Page();
    ko.applyBindings(window.page);
  };
}).call(window);
