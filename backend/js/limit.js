export default function () {
    let data={}
    if (options && options.size > 0) {
        data.qty = options.size
    }
    if (options.price > 0) {
        data.price = options.price
    }
    if (options.limit > 0) {
        data.limit = options.limit
    }
    data.comment= "止赢"
    console.log("[limit-1]",options)
    let res = ex.entry("algo", options.side, data)
    console.log("[limit-2]",res)
    return res
}