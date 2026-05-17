var s = """# Markdown Feature Showcase

A document demonstrating **all major Markdown features**.

---

## 1. Headings

```markdown
# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6
```

# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6

---

## 2. Text Formatting

| Feature | Syntax | Rendered |
|---------|--------|----------|
| **Bold** | `**bold**` or `__bold__` | **bold** |
| *Italic* | `*italic*` or `_italic_` | *italic* |
| ***Bold+Italic*** | `***bold+italic***` | ***bold+italic*** |
| ~~Strikethrough~~ | `~~strikethrough~~` | ~~strikethrough~~ |
| `Inline code` | `` `code` `` | `Inline code` |
| **<u>Underline</u>** | HTML `<u>` tag | <u>Underline</u> |
| **SUP** | `<sup>` tag | <sup>SUP</sup> |
| **SUB** | `<sub>` tag | <sub>SUB</sub> |

---

## 3. Links

- **Auto-link**: <https://example.com>
- **Inline link**: [Example](https://example.com)
- **Link with title**: [Example](https://example.com "Visit Example")
- **Reference link**: [Example][ex-link]
- **Image link**: [![Alt text](https://via.placeholder.com/150x50)](https://example.com)
- **Email link**: <user@example.com>

[ex-link]: https://example.com "Example Website"

---

## 4. Images

![Markdown Logo](https://upload.wikimedia.org/wikipedia/commons/4/48/Markdown-mark.svg)

*Figure 1: The Markdown logo.*

---

## 5. Blockquotes

> This is a single-level blockquote.
>
> > This is a nested blockquote (level 2).
> >
> > > This is level 3.

> **Note:** You can include **bold**, *italic*, `code`, and even [links](https://example.com) inside blockquotes.

> ### Quoted heading
> This is a heading inside a blockquote.

---

## 6. Lists

### Unordered List

- First item
- Second item
  - Nested item A
  - Nested item B
    - Deeply nested
- Third item
  - **Bold item**
  - *Italic item*
  - `Code item`

### Ordered List

1. First item
2. Second item
3. Third item
   1. Nested ordered A
   2. Nested ordered B

### Task List

- [x] Completed task
- [ ] Incomplete task
- [ ] Another incomplete task
- [x] Yet another completed task

### Definition List (HTML)

<dl>
  <dt>Term 1</dt>
  <dd>Definition of term 1</dd>
  <dt>Term 2</dt>
  <dd>Definition of term 2</dd>
</dl>

---

## 7. Code Blocks

### Fenced Code Block (no language)

```
function hello() {
  console.log("Hello, world!");
}
```

### Fenced Code Block (with syntax highlighting)

```python
def fibonacci(n: int) -> list[int]:
    seq = [0, 1]
    for _ in range(2, n):
        seq.append(seq[-1] + seq[-2])
    return seq[:n]

print(fibonacci(10))
```

```javascript
// Arrow function with destructuring
const greet = ({ name, greeting = "Hello" }) =>
  `${greeting}, ${name}!`;

console.log(greet({ name: "Alice" }));
// → "Hello, Alice!"
```

```bash
#!/bin/bash
# Install dependencies
go mod tidy
go build -o blue ./cmd/blue
./blue --version
```

### Indented Code Block

    This is a code block
    created with 4 spaces of indentation.
    It doesn't need backticks.

---

## 8. Tables

### Basic Table

| Feature        | Syntax                  | Description              |
|----------------|-------------------------|--------------------------|
| Heading        | `## Heading`            | Creates a section title  |
| Link           | `[text](url)`           | Creates a hyperlink      |
| Image          | `![alt](url)`           | Embeds an image          |
| Code block     | ```` ``` ````           | Fenced code block        |
| Table          | `| col1 \| col2 \|`     | Creates a table          |

### Table with Alignment

| Left-aligned | Centered | Right-aligned |
|:-------------|:--------:|--------------:|
| Item 1       |   Item 2 |        Item 3 |
| Item 4       |   Item 5 |        Item 6 |
| Item 7       |   Item 8 |        Item 9 |

### Table with Inline Formatting

| Name | Status | Notes |
|------|--------|-------|
| **Parser** | ✅ Done | Recursive descent |
| *Compiler* | ✅ Done | Bytecode generation |
| `VM` | ✅ Done | Stack-based execution |

---

## 9. Horizontal Rules

Three or more hyphens:

---

Three or more asterisks:

***

Three or more underscores:

___

---

## 10. Line Breaks

This is line 1.  
This is line 2 (note the two spaces at the end of line 1).

This is line 3.
<br>
This is line 4 (using HTML `<br>` tag).

---

## 11. Escaping Characters

To include a literal backtick: \`\`\`

To include a literal asterisk: \*not italic\*

To include a literal hash: \# not a heading

To include a literal bracket: \[not a link\]

---

## 12. HTML in Markdown

<div style="background-color: #f0f0f0; padding: 10px; border-left: 4px solid #4a90d9;">
  <strong>HTML block:</strong> You can embed raw HTML inside Markdown.
</div>

<span style="color: red;">This text is red.</span>
<span style="color: green;">This text is green.</span>
<span style="color: blue;">This text is blue.</span>

<details>
<summary>Click to expand this details section</summary>

This content is **hidden** by default and revealed when clicked.

- You can include lists
- You can include `code`
- You can include **formatted text**

</details>

---

## 13. Footnotes

Here is a sentence with a footnote[^1]. Here is another with multiple references[^2][^3].

[^1]: This is the first footnote. It can contain **bold**, *italic*, and even [links](https://example.com).
[^2]: This is the second footnote. It demonstrates that footnotes can have different lengths.
[^3]: This is the third footnote, proving you can have many footnotes in a single document.

---

## 14. Special Elements

### Abbreviation

The HTML specification is maintained by the W3C.[^abbr]

[^abbr]: World Wide Web Consortium

### Inline HTML Attributes (in some parsers)

Here is some text with a custom attribute. {data-custom="value"}

---

## 15. Math (via LaTeX in some parsers)

Inline math: $E = mc^2$

Block math:

$$
\int_{-\infty}^{\infty} e^{-x^2} dx = \sqrt{\pi}
$$

---

## 16. Emoji

Here are some popular emoji: 🎉 🚀 💻 🎨 🔥 ⭐ 🌟 💯 ✅ ❌ 🎯

---

## 17. Mermaid Diagrams (in some parsers)

```mermaid
graph TD
    A[Source Code] --> B[Lexer]
    B --> C[Parser]
    C --> D[AST]
    D --> E{Compiler}
    D --> F{Evaluator}
    E --> G[Bytecode]
    F --> H[Direct Execution]
    G --> I[VM]
```

---

## 18. PlantUML Diagrams (in some parsers)

```plantuml
@startuml
actor User
participant "Blue CLI" as CLI
participant "Compiler" as C
participant "VM" as VM

User -> CLI: blue program.b
CLI -> C: compile
C -> VM: run bytecode
VM --> User: output
@enduml
```

---

## 19. Checklist Comparison

| Feature | Supported | Notes |
|---------|-----------|-------|
| Headings | ✅ | 6 levels |
| Bold / Italic | ✅ | Multiple syntaxes |
| Strikethrough | ✅ | Requires GFM |
| Links | ✅ | Inline, reference, auto |
| Images | ✅ | Inline, reference |
| Blockquotes | ✅ | Nested |
| Lists | ✅ | Ordered, unordered, task |
| Code blocks | ✅ | Fenced, indented, highlighted |
| Tables | ✅ | With alignment |
| Horizontal rules | ✅ | Multiple styles |
| HTML | ✅ | Inline and block |
| Footnotes | ✅ | GFM extension |
| Emoji | ✅ | Inline |
| Math (LaTeX) | ⚠️ | Parser-dependent |
| Mermaid | ⚠️ | Parser-dependent |
| PlantUML | ⚠️ | Parser-dependent |

---

*Document generated to showcase Markdown features.*
*Some features (math, mermaid, plantuml) depend on the rendering engine.*
""";

import http

val s1 = http.md_to_html(s);
var s2 = http.html_to_md(s1);
println("---------------------------------------------------------------------------------------------------------------");
println(s);
println("---------------------------------------------------------------------------------------------------------------");
println(s1);
println("---------------------------------------------------------------------------------------------------------------");
println(s2);
println("---------------------------------------------------------------------------------------------------------------");
# The translations are not 1:1 but its close, so just basic smoke test added below
assert(http.md_to_html('# Hello World') == '<h1>Hello World</h1>\n');
assert(http.html_to_md('<h1>Hello World</h1>') == '# Hello World')