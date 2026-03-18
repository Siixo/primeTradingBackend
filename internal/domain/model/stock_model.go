package model

type Stock struct {
	Commodity string 
	Date string `json:"gmt_ounce_price_usd_updated"`
	Price float32 `json:"kg_in_usd"`
}


/* "{\"ounce_price_usd\":\"5006.270\"
,\"gmt_ounce_price_usd_updated\":\"16-03-2026 09:29:46 pm\",
\"ounce_price_ask\":\"5006.270\",
\"ounce_price_bid\":\"5006.270\",
\"ounce_price_usd_today_low\":\"4967.93\",
\"ounce_price_usd_today_high\":\"5038.04\",
\"usd_to_eur\":\"0.868953121\",
\"gmt_eur_updated\":\"16-03-2026 09:28:59 pm\",
\"ounce_in_eur\":4350.213941068670465028844773769378662109375,\"gram_to_ounce_formula\":0.032099999999999996591615314400769420899450778961181640625,
\"gram_in_usd\":160.701267000000001416992745362222194671630859375,
\"gram_in_eur\":139.64186750830430128189618699252605438232421875}"% 

"{\"ounce_price_usd\":\"5006.270\",
\"gmt_ounce_price_usd_updated\":\"16-03-2026 09:35:47 pm\",
\"ounce_price_ask\":\"5006.270\",
\"ounce_price_bid\":\"5006.270\",

\"ounce_price_usd_today_low\":\"4967.93\",
\"ounce_price_usd_today_high\":\"5038.04\",
\"kg_to_ounce_formula\":32.150700000000000500222085975110530853271484375,
\"kg_in_usd\":160955.08488900001975707709789276123046875}"%   */