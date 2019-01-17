# Gas

## Gas项

1. Store

| 操作 | Gas值 | 说明 |
| :--- | :---: | :--- |
| HasCost |          10 |   |
| DeleteCost |       10 |   |
| ReadCostFlat |     10 |   |
| ReadCostPerByte |  1 |    |
| WriteCostFlat |    10 |   |
| WriteCostPerByte | 10 |   |
| ValueCostPerByte | 1 |    |
| IterNextCostFlat | 15 |   |

2. ITx

ITx实现类自定义`CalcGas()`，针对不同Tx收取不同的Gas。

## 实现

`BaseApp`定义`GasMeter`来实时计算Gas消耗，`GasHandler`供上层应用实现`Gas`的处理逻辑。

### InitChain

使用`InfiniteGasMeter`运行创世块。

### CheckTx

根据参数`MaxGas`和`CalcGas()`设置`GasMeter`，完整执行`Tx`逻辑后执行`GasHandler()`，`Gas`不足时实时返回相关错误信息。

### DeliverTx

同[CheckTx](#CheckTx)，不同之处在于错误信息会保存到区块中。