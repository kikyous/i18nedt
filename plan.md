我要做一个在命令行运行的i18n编辑器，i18nedt

* 语言 go
* 命令行运行

## 用法
i18nedt src/locales/{zh-CN,zh-TW,en-US}.json -k home.welcome

会打开vim（$editor），并创建一个临时文件，格式如下：

```i18nedt-${timestamp}.txt
# home.welcome
* zh-CN

* zh-TW

* en-US

```
用户可以编辑这个临时文件。

```i18nedt-${timestamp}.txt
# home.welcome
* zh-CN
欢迎

* zh-TW
歡迎

* en-US
Welcome
```

当编辑器退出时，写入对应的json文件

## 删除key

```i18nedt-${timestamp}.txt
#- home.welcome
# home.welcomeNew
* zh-CN
欢迎

* zh-TW
歡迎

* en-US
Welcome
```

#后紧跟-的key会被删除

## 说明
-k 参数可以有多个，如：
i18nedt src/locales/{zh-CN,zh-TW,en-US}.json -k home.welcome -k home.start
