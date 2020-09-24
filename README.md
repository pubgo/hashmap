# hashmap
hashmap for go

优化的点
1. 快速的hash函数
2. 快速查找定位key的方法
3. 减少对象内存分配次数
4. 减少迁移的量

## 链桶
1. 先分桶，然后每个桶都是链
2. 迁移方便，迁移就是数据的移动
3. 查询需要遍历，有点慢，并不能内存加速
4.

## 数组桶
1. 先分桶，然后每个桶里面是数组，顺序存放数据
2. 迁移不方便，还是需要数组的移动

## 数组+链桶
1. 先分桶
2. 桶里存放8个，重复的，放到overflow链中
