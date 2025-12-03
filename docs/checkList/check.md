## 🧪 Code Review Checklist

> ✅ 请 reviewer 逐项确认；作者应在提交前自检

### 1️⃣ 功能正确性（Correctness）
- [ ] 逻辑符合需求，无功能偏差或遗漏
- [ ] 边界条件（空值、极值、异常输入）已处理
- [ ] 新增/修改代码有对应单元测试或集成测试
- [ ] 未引入回归问题（与历史行为一致）

### 2️⃣ 可读性与可维护性（Readability & Maintainability）
- [ ] 命名清晰、语义化（变量/函数/类）
- [ ] 函数职责单一，长度合理（建议 ≤ 50 行）
- [ ] 无重复或冗余代码（遵循 DRY）
- [ ] 关键逻辑有注释或文档说明（非“废话注释”）

### 3️⃣ 健壮性与错误处理（Robustness）
- [ ] 异常路径有处理（try/catch、fallback、重试等）
- [ ] 外部调用（API/DB/文件）设超时、限流或降级
- [ ] 无“静默失败”——错误应被记录或上报
- [ ] 资源（文件、连接、锁）正确释放

### 4️⃣ 安全性（Security）
- [ ] 无硬编码密钥、密码、Token 🔑
- [ ] 用户输入经过校验、转义或参数化（防 XSS/SQLi）
- [ ] 敏感数据未明文打印到日志或返回前端
- [ ] 权限校验在服务端完成（非仅前端）

### 5️⃣ 编码规范与一致性（Style & Consistency）
- [ ] 无硬编码
- [ ] 符合团队代码风格（缩进、命名、格式）
- [ ] 未使用已废弃（deprecated）的 API 或库
- [ ] 依赖版本合理，未引入不必要新包
- [ ] Git 提交信息清晰，关联 Issue / Ticket


---
给AI审核用的结构化规则清单（JSON Schema 风格，便于解析）
```
{
  "review_rules": [
    {
      "category": "correctness",
      "name": "logic_matches_requirement",
      "description": "Code logic aligns with functional requirements",
      "severity": "high",
      "ai_checkable": true,
      "evidence_type": ["diff_analysis", "test_coverage"]
    },
    {
      "category": "correctness",
      "name": "boundary_conditions_handled",
      "description": "Edge cases (null, empty, extreme values) are handled",
      "severity": "medium",
      "ai_checkable": true,
      "evidence_type": ["static_analysis", "pattern_matching"]
    },
    {
      "category": "readability",
      "name": "meaningful_naming",
      "description": "Variables, functions, and classes use clear, semantic names",
      "severity": "medium",
      "ai_checkable": true,
      "evidence_type": ["naming_heuristic", "entropy_check"]
    },
    {
      "category": "readability",
      "name": "function_length",
      "description": "Function length <= 50 lines",
      "severity": "low",
      "ai_checkable": true,
      "evidence_type": ["ast_parsing"],
      "threshold": 50
    },
    {
      "category": "robustness",
      "name": "exception_handling",
      "description": "External calls or risky operations have try/catch or error handling",
      "severity": "high",
      "ai_checkable": true,
      "evidence_type": ["ast_parsing", "call_graph_analysis"]
    },
    {
      "category": "security",
      "name": "no_hardcoded_secrets",
      "description": "No hardcoded passwords, API keys, or tokens in source",
      "severity": "critical",
      "ai_checkable": true,
      "evidence_type": ["regex_scan", "secret_pattern_db"]
    },
    {
      "category": "security",
      "name": "input_sanitization",
      "description": "User inputs are validated, escaped, or parameterized",
      "severity": "high",
      "ai_checkable": true,
      "evidence_type": ["data_flow_analysis", "taint_tracking"]
    },
    {
      "category": "style",
      "name": "follows_style_guide",
      "description": "Code conforms to team linter/formatter rules",
      "severity": "low",
      "ai_checkable": true,
      "evidence_type": ["linter_output"]
    },
    {
      "category": "style",
      "name": "no_deprecated_apis",
      "description": "Does not use deprecated functions or libraries",
      "severity": "medium",
      "ai_checkable": true,
      "evidence_type": ["dependency_analysis", "api_catalog"]
    }
  ],
  "scoring_weights": {
    "correctness": 0.30,
    "readability": 0.20,
    "robustness": 0.20,
    "security": 0.20,
    "style": 0.10
  },
  "severity_levels": {
    "critical": -10,
    "high": -5,
    "medium": -2,
    "low": -1
  }
}
```