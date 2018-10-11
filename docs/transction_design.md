# <font color=#0099ff>Transaction设计概要</font>
![tx_struct](https://github.com/QOSGroup/static/blob/master/transaction_struct.jpg?raw=true)  
	Transaction（下文记为 tx）设计时主要考虑几种应用情景：  
	1，跨链传输tx，对应结构体：TxQcp  
	2，链内标准tx, 对应结构体：TxStd  
	3，具体功能tx, 对应结构体：参考"qos/transaction_design_qos.md"
# <font color=#0099ff>TxStd & ITx</font>
	公链提供几种基础的 tx，每种 tx 有自己的结构和执行逻辑：  

        1，创建 QSC，   结构：TxCreateQSC  
        2，发币，       结构：TxIssueQsc  
        3，普通交易，   结构：TxTransform  
        4，Qcp执行结果，结构：QcpTxReasult  

	将上述几种基础 tx 抽象为接口 ITx，定义标准结构TxStd，接口ITx作为TxStd的成员，从而使不同的基础tx呈现为统一的结构（TxStd），接口中定义了tx的基本操作，每种基础tx需做相应实现。  
	
	ITx中定义了功能tx需要实现的接口：
```
	type ITx interface{
		ValidateData()   //检测
		Exec()           //执行
		GetSigner()      //签名者
		CalcGas()        //计算gas
		GetGasPayer()    //gas付费人
		GetSignData()    //获取签名字段
	}
```
# <font color=#0099ff>TxQcp</font>
跨链的tx需要更多信息，其结构 TxQcp 中除TxStd结构（Payload成员）外，还含有以下成员：  

		From：		qsc的名字，描述来源
		To：		    qos名字，描述目的地
		Sequence：	序号，避免重复，也用于中继传输和执行结果对应
		IsResult：	是否为执行结果，避免执行结果被作为 TxQcp 再次发送
		BlockHeight: Tx所在block的高度
		TxIndx：     Tx在block中的位置
# <font color=#0099ff>QcpTxResult</font>
功能：TxQcp在公链上的执行结果；

	Code：	 执行结果(成功、失败等)
	Extends：扩展信息，方便结果传递更多的内容
	GasUsed：gas消耗值
	QcpOriginalSequence：指示对应的 TxQcp
	Info：	描述信息（如：错误信息等）

此Tx为公链上的执行结果，其封装为TxQcp结构（TxQcp.IsResult = True）后传至联盟链执行，联盟链执行后的结果不会再次封装成 TxQcp 发往公链（否则陷入循环）。