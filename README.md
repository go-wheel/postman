# postman
Batch HTTP requests, using postman format


1. 命令行参数
taskpath: http任务地址，postman格式描述
resultpath: 处理结果
rangetpl:断点上传使用模板

2. 机制
http任务直接往taskpath目录下增加，程序会主动获取新增文件处理任务。
断点上传的任务在%taskpath%/range目录下，断点上传接口比较复杂，目前是定制化开发，需要模板加数据分离方式
