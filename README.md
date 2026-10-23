# in3words

in3words 是[三言三语](https://best33.com)的前端项目。

in3words 使用 React Router（framework 模式）用于 SSR，后端使用 WordPress 作为 headless CMS，并希望在将来移除 WordPress，成为一个独立的 CMS 项目。

## server

`server/` 是用 Go 编写的 WordPress REST API 只读替代实现，直接读取 WordPress 的 MySQL 数据库，提供前端用到的站点信息、文章、页面、分类、标签、用户与评论读取接口（路径与 `/wp-json/` 一致），以及全文 RSS（`/feed/`）。发表评论仍需 WordPress。

```sh
cd server
cp config.example.yaml config.yaml # 按 wp-config.php 填写数据库信息
go run . serve -c config.yaml
```

配置项也可以用环境变量覆盖，例如 `I3W_DB_PASSWORD`。

## 开源许可

程序代码，含 ts, tsx, js, css, scss 等纯文本文件，以 BSD-3-Clause 授权：

```
For code in this repository, files ending with .ts, .tsx, .js, .css, .scss and other plain text files:

Copyright 2021 oott123

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```

其它设计资源，jpg, png 等图片文件，保留所有权利，不授权：

```
For all other files:

Copyright 2021 oott123 All rights reserved.
```
