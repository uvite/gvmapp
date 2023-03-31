export default function () {


        let qty=1
        console.log("[333]",options.size)
        if(options&&options.size){
                qty=options.size
        }

        let res=strategy.exit("long")
        console.log(res)
        return res
}