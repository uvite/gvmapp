import {ref} from "vue";
import {useTimeout} from "v3hooks";

var _assign = function __assign() {
    _assign = Object.assign || function __assign(t) {
        for (var s, i = 1, n = arguments.length; i < n; i++) {
            s = arguments[i];

            for (var p in s) {
                if (Object.prototype.hasOwnProperty.call(s, p)) t[p] = s[p];
            }
        }

        return t;
    };

    return _assign.apply(this, arguments);
};
var ReadyState;
(function (ReadyState) {
    ReadyState[ReadyState["Connecting"] = 0] = "Connecting";
    ReadyState[ReadyState["Open"] = 1] = "Open";
    ReadyState[ReadyState["Closing"] = 2] = "Closing";
    ReadyState[ReadyState["Closed"] = 3] = "Closed";
})(ReadyState || (ReadyState = {}));
var defaultOptions = {
    manual: false,
    reconnectLimit: 3,
    reconnectInterval: 3000,
    onOpen: function () { },
    onClose: function () { },
    onMessage: function () { },
    onError: function () { },
};
function useSocket(socketUrl, options) {
    var _a = _assign(_assign({}, defaultOptions), options), manual = _a.manual, reconnectLimit = _a.reconnectLimit, reconnectInterval = _a.reconnectInterval, onOpen = _a.onOpen, onClose = _a.onClose, onMessage = _a.onMessage, onError = _a.onError;
    if (!socketUrl || typeof (socketUrl) !== 'string') {
        throw new Error('useWebSocket require string socketUrl');
    }
    var readyState = ref(ReadyState.Connecting);
    var reconnectCount = ref(0);
    var socket = ref();
    var latestMessage = ref();
    var run = function () {
        socket.value = new WebSocket(socketUrl);
        socket.value.addEventListener('open', function (event) {
            readyState.value = ReadyState.Open;
            onOpen(event);
        });
        socket.value.addEventListener('message', function (event) {
            latestMessage.value = event;
            onMessage(event);
        });
        socket.value.addEventListener('error', function (event) {
            console.log('error ', event);
            reconnect();
            onError(event);
        });
        socket.value.addEventListener('close', function (event) {
            readyState.value = ReadyState.Closed;
            onClose(event);
        });
    };
    var connect = function () {
        if (readyState.value !== ReadyState.Open) {
            reconnectCount.value = 0;
            run();
        }
    };
    var reconnect = function () {
        if (reconnectCount.value >= reconnectLimit)
            return;
        useTimeout(function () {
            reconnectCount.value++;
            run();
        }, ref(reconnectInterval));
    };
    var disconnect = function () {

        if ((readyState.value === ReadyState.Connecting
                || readyState.value === ReadyState.Open)
            && socket.value) {
            console.log("[q1]",ReadyState.Connecting)
            readyState.value = ReadyState.Closing;
            let a=socket.value.close();
            console.log("[q2]",a)
        }
    };
    var sendMessage = function (data) {
        if (data
            && socket.value
            && readyState.value === ReadyState.Open)
            socket.value.send(data);
    };
    if (!manual)
        connect();
    return {
        latestMessage: latestMessage,
        readyState: readyState,
        connect: connect,
        disconnect: disconnect,
        sendMessage: sendMessage,
        webSocketIns: socket
    };
}
export default useSocket
