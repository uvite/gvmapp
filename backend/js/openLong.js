export default function (data) {
    let qty = 1

    if (options && options.size > 0) {
        qty = options.size
    }

    let res = strategy.entry("buy", "Long", {qty: qty.toString(), comment: "开多"})
    console.log("[22]",res)
    return res
}