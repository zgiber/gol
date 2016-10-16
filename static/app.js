var canvas = document.getElementById('myCanvas'); 
var ctx = canvas.getContext('2d');
var newPoints = [];
var drawTimeout;

function constructImageData(points) {
    var data = new Uint8ClampedArray(600 * 600 * 4);
    var pointsLength = points.length;

    for (var i = 0; i < pointsLength; i++) {
        var point = points[i]
        var position = (point.x + (600 * point.y)) * 4;
        data[position] = 255;
        data[position + 3] = 255;
    }

    return new ImageData(data, 600, 600)
}

var socket = new WebSocket("ws://" + window.location.host + "/ws");

socket.onmessage = function (event) {
    var points = JSON.parse(event.data);
    ctx.putImageData(constructImageData(points), 0, 0);
}

function sendPoints() {
    if (length.newPoints == 0) {
        return
    }

    var pointData = JSON.stringify(newPoints);
    socket.send(pointData);
    newPoints = [];
}

function drawPoint(x, y) {
    clearTimeout(drawTimeout);
    var newPoint = { x: x, y: y };
    newPoints.push(newPoint);
    drawTimeout = setTimeout(sendPoints, 100);
}

function draw(evt) {
    var mousePos = getMousePos(evt);
    drawPoint(mousePos.x, mousePos.y);
}

function getMousePos(evt) {
    var rect = canvas.getBoundingClientRect();
    return {
        x: evt.clientX - rect.left,
        y: evt.clientY - rect.top
    };
}
