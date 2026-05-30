[English](./README.md) / **中文**

---

**PlainMem，一个面向 AI Agent 的长期记忆系统。** 可移植、纯文件存储、零依赖、不绑定任何框架。

安装后运行 `pmem init` 即可。Agent 从此拥有跨会话记忆。

灵感来源于 [GenericAgent](https://github.com/lsdefine/GenericAgent) 的记忆架构。

---

## 架构

### 记忆层次

```
┌─────────────────────────────────────────────────────────────┐
│  L0: 宪法 (内嵌在源代码中)                                    │
│  • 不可修改的系统规则                                         │
│  • 任何 Agent 都无法修改                                     │
│  • 由 MemorySystem 强制执行                                  │
└─────────────────────────────────────────────────────────────┘
                              ↓ 执行约束
┌─────────────────────────────────────────────────────────────┐
│  L1: 洞察索引 (global_mem_insight.txt)                       │
│  • ≤30行，指针式路由                                          │
│  • [RULES] 部分用于行为指南                                   │
│  • L2/L3 分类引用                                            │
│  • 密钥检测 & 行数验证                                        │
└─────────────────────────────────────────────────────────────┘
                              ↓ 指向
┌─────────────────────────────────────────────────────────────┐
│  L2: 事实存储 (global_mem.txt)                               │
│  • 环境事实 (路径、配置、API)                                 │
│  • 按分类组织: [env], [config], [decision]                    │
│  • 自动与 L1 索引同步                                         │
└─────────────────────────────────────────────────────────────┘
                              ↓ 引用
┌─────────────────────────────────────────────────────────────┐
│  L3: 程序记忆 (pm/*.md, *.py)                                │
│  • 可复用的 SOP 和脚本                                       │
│  • 索引维护在 sop_index.json                                 │
│  • 模板: 定义、工作流、红线                                    │
└─────────────────────────────────────────────────────────────┘
```

### 核心设计原则

1. **分层架构**: 清晰的关注点分离 (L0 不可变规则 → L3 可复用程序)
2. **文件存储**: 纯文本文件，无数据库依赖，完全可移植
3. **索引同步**: L2/L3 变化时 L1 自动更新
4. **不可变宪法**: L0 存在于源代码中，Agent 无法修改
5. **验证优先**: 存储前验证，写入前搜索
6. **MCP 兼容**: 标准协议，支持任何 AI 主机集成

---

## 快速开始

```bash
# 从 Releases 下载二进制，或从源码构建：
git clone https://github.com/ThinkMars/plain-mem.git
cd plain-mem
go build -o ~/.local/bin/pmem ./cmd/pmem/

# 初始化 — 自动将记忆配置写入 AGENTS.md：
pmem init
```

完成。AGENTS.md 会自动告诉你的 Agent 如何使用 pmem。

---

## 命令

```bash
pmem read                    # 加载记忆到提示词（会话开始时执行）
pmem agents                  # 打印完整使用指令
pmem init                    # 将记忆配置写入 AGENTS.md

pmem add-fact <分类> <内容>  # 添加事实（section 存在则追加）
pmem append-fact <分类> <内容> # 追加行到已有 section
pmem update-fact <分类> <内容> # 替换整个 section（原地 patch）

pmem add-rule <文本>         # 添加行为规则（L1）

pmem add-sop <名称> [标题]   # 创建 SOP 文件（L3）
pmem update-sop <名称> <内容> # 替换 SOP 内容
pmem list-sops               # 列出所有 SOP

pmem search <关键词>         # 搜索 L2 和 L3
pmem consolidate             # 输出记忆沉淀指令
pmem verify                  # 检查记忆完整性
pmem mcp                     # 以 MCP 服务器模式运行（stdio）
```

多行内容通过 stdin 传入（使用 `-`）：

```bash
pmem update-fact env - <<'EOF'
python: 3.13
node: 22
EOF
```

---

## 示例

```bash
# 加载记忆
pmem read

# 存储事实（section 存在则自动追加）
pmem add-fact env "python: /opt/homebrew/bin/python3.12"
pmem add-fact config "port 8080, db sqlite"

# 添加规则
pmem add-rule "发起外部API请求时必须设置超时。"

# 创建和更新 SOP
pmem add-sop deploy "部署检查清单"
pmem update-sop deploy - <<'EOF'
# 部署检查清单
## 工作流
1. go test ./...
2. go build -o pmem ./cmd/pmem/
3. pmem verify
EOF
pmem list-sops

# 搜索
pmem search python

# 会话结束
pmem consolidate
pmem verify
```

---

## 存储

记忆以纯文本文件存储在 `./pm/` 下：

```
pm/
├── global_mem_insight.txt      # L1 索引
├── global_mem.txt              # L2 事实
├── sop_index.json              # L3 元数据
└── deploy_sop.md               # L3 SOP文件
```

---

## License

[MIT](https://github.com/ThinkMars/plain-mem/blob/main/LICENSE)
