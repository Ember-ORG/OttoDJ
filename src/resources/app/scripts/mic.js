var mic = true;
var first = true;
var constraints = {
    audio: {
        echoCancellation: true,
        noiseSuppression: false,
        autoGainControl: true,
    }
};
var output = document.getElementById("audio");
var micimg = document.getElementById("micimg");

function micToggle() {
    if (first) {
        navigator.mediaDevices.getUserMedia(constraints)
            .then(stream => output.srcObject = stream)
            .catch(e => log(e));
        var log = msg => div.innerHTML += msg + "<br>";
        first = false;
    }

    if (mic) {
        mic = false;
        output.play();
        micimg.innerHTML = "mic_on";
    } else {
        mic = true;
        output.pause()
        micimg.innerHTML = "mic_off";
    }
}