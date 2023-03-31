export default function () {
    let data={}
    if (options && options.size > 0) {
        data.qty = options.size
    }
    if (options.price > 0) {
        data.price = options.price
    }
    if (options.stop > 0) {
        data.stop = options.stop
    }
    data.comment= "止赢"
    console.log("[11]",options)
    let res = ex.entry("algo", options.side, data)
    console.log("[22]",res)
    return res
}