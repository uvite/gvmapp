export default function (data) {
   let  trades  = ex.getTrades()
   console.log(trades)
   let balance=ex.getBalance()
   console.log(balance)
   let postion=ex.getPostions()
   console.log(postion)
}