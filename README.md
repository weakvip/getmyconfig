# getmyconfig
A VERY SIMPLE HTTP SERVER FOR DISTRIBUTION CONFIG, NO TECH JUST FOR FUN

分布式配置服务端
数据为JSON格式, 相同KEY首次GET时保存为默认值, POST更新默认值
URL表单内NAME为配置名称, VALUE为JSON格式, 传入JSON属性数量不固定(目前仅支持一层JSON)
例如: http://localhost:8080/?conf={"k1":"v1","k2":"v2"}&conf1={"k":123}
功能单一，没有什么实用价值仅作为练手
