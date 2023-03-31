import WebSocket from 'ws';

var last_event = Infinity
var token = ''
var ws = null
var _ontrades = () => {}
var _onquotes = () => {}
var _onready = () => {}
var _onrefine = () => {}
var reconnecting = false
var ready = false
var symbols = []
var terminated = false

function now() {
    return new Date().getTime();
}

async function init(syms) {

    if (ready) return

    symbols = syms
    start_hf(symbols)

    // If connection error, try again
    setTimeout(() => init(symbols), 10000);
}

function start_hf() {
    // To subscribe to this channel:
    var msg = syms => ({
        'method': 'SUBSCRIBE',
        "params": syms,
        "id": 1
    })

    ws = new WebSocket(`wss://stream.binance.com:443/ws`);
    ws.onmessage = function(e) {
        try {
            // console.log('Server data', e.data);
            let data = JSON.parse(e.data);
            if (!data.s) return print(data)
            switch (data.e) {
                case 'aggTrade':
                    _ontrades({
                        symbol: data.s.toUpperCase(),
                        price: parseFloat(data.p),
                        size: parseFloat(data.q),
                    })
                    break
                case 'kline':
                    _ontrades({
                        symbol: data.s.toUpperCase(),
                        price: parseFloat(data.p),
                        size: parseFloat(data.q),
                    })
                    break
                case 'ping':
                    console.log('PING', data)
                    break
            }
            last_event = now()
        } catch (e) {
            console.log(e.toString())
        }
    };
    ws.onopen = function() {
        try {
            let syms = symbols.map(x =>
                x.toLowerCase() + "@aggTrade")
            console.log('SEND >>>', JSON.stringify(msg(syms)))
            ws.send(JSON.stringify(msg(syms)))
        } catch(e) {
            console.log(e.toString())
        }
    };
    ws.onclose = function (e) {
        switch (e) {
            case 1000:
                console.log("WebSocket: closed");
                break;
        }
        //reconnect();
    };
    ws.onerror = function (e) {
        console.log("WS", e);
        reconnect();
    };
}

function add_symbol(s) {
    var msg = sym => ({
        'method': 'SUBSCRIBE',
        'channel': 'trades',
        'market': sym.toLowerCase()+"@aggTrade"
    });
    try {
        ws.send(JSON.stringify(msg(s)));
        symbols = [s];
        ws.close();
        start_hf(symbols);
    } catch(e) {
        console.log(e.toString())
    }
}
function remove_symbol(s) {
    let msg = sym => ({
        'method': 'UNSUBSCRIBE',
        'channel': 'trades',
        'market': sym.toLowerCase()+"@aggTrade"
    });
    try {
        ws.send(JSON.stringify(msg(s)))
        symbols = [];
    } catch(e) {
        console.log(e.toString())
    }
}

function reset(s) {
    let remove = sym => ({
        'method': 'UNSUBSCRIBE',
        'channel': 'trades',
        'market': sym.toLowerCase()+"@aggTrade"
    });
    try {
        ws.send(JSON.stringify(remove(symbols[0])))
        symbols = [];
    } catch(e) {
        console.log(e.toString())
    }

    var msg = sym => ({
        'method': 'SUBSCRIBE',
        'channel': 'trades',
        'market': sym.toLowerCase()+"@aggTrade"
    });
    try {
        ws.send(JSON.stringify(msg(s)));
        symbols = [s];
        ws.close();
        start_hf(symbols);
    } catch(e) {
        console.log(e.toString())
    }
}

function reconnect() {
    reconnecting = true
    console.log('Reconnecting...')
    try {
        ws.close()
        setTimeout(() => start_hf(symbols) , 1000)
    } catch(e) {
        console.log(e.toString())
    }
}

function print(data) {
    // TODO: refine the chart
    if (reconnecting) {
        _onrefine()
    } else if (!ready) {
        console.log('Stream [OK]')
        _onready()
        ready = true
        last_event = now()
        setTimeout(heartbeat, 10000)
    }
    reconnecting = false
}

function heartbeat() {
    if (terminated) return
    if (now() - last_event > 60000) {
        console.log('No events for 60 seconds')
        if (!reconnecting) reconnect()
        setTimeout(heartbeat, 10000)
    } else {
        setTimeout(heartbeat, 1000)
    }
}

function terminate() {
    ws.close()
    console.log('Stream [Close]');
    terminated = true
}

export default {
    init,
    add_symbol,
    reconnect,
    terminate,
    remove_symbol,
    reset,
    set ontrades(val) {
        _ontrades = val
    },
    set onquotes(val) {
        _onquotes = val
    },
    set ready(val) {
        _onready = val
    },
    set refine(val) {
        _onrefine = val
    },

}
