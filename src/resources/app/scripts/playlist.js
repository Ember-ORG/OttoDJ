//Defining variables
var songSelected = false;

document.getElementById("file_input").onchange = function(e) {
    if (document.getElementById("file_input").files.length != 0) {
        songSelected = true;
    }
    $("#play-button")
        .addClass("play-active")
        .addClass("complete")
        .removeClass("play-inactive")
        .removeClass("unchecked")
        .removeClass("icon-play")
        .addClass("icon-one");
    $("#play-button")
        .children(".icon")
        .removeClass("icon-play")
        .addClass("icon-cancel");
    $(".info-two")
        .addClass("info-active");
    $("#pause-button")
        .addClass("scale-animation-active");
    $(".waves-animation-one, #pause-button, .seek-field, .volume-icon, .volume-field, .info-two").show();
    $(".waves-animation-two").hide();
    $("#pause-button")
        .children('.icon')
        .addClass("icon-pause")
        .removeClass("icon-play");
    setTimeout(function() {
        $(".info-one").hide();
    }, 400);
    var message
    for (var i = 0; i < this.files.length; i++) {
        if (i == 0) {
            message = this.files[i].path
        } else {
            message = message + "," + this.files[i].path
        }
    }
    astilectron.sendMessage("p" + message, function(message) {
        console.log("received " + message)
    });
}