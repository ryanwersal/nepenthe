---
name: review
description: Review code changes for correctness, security, concurrency, and project conventions
argument-hint: [diff-target]
context: fork
agent: nepenthe-reviewer
allowed-tools: Bash, Read, Grep, Glob
---

# Review Changes: $ARGUMENTS

## Step 1: Understand the Changes

Run the following to understand what changed:

```
git diff $ARGUMENTS --stat
git diff $ARGUMENTS
git log $ARGUMENTS --oneline
```

Then read every modified file in full to understand surrounding context. Do not review lines in isolation — understand the function, the file, and how it fits the architecture.

## Step 2: Apply the Review Checklist

Review every changed line against ALL categories in your review checklist. Be thorough.

## Step 3: Check for Anti-Patterns

Scan all changed code for the 14 anti-patterns in your checklist. Flag every occurrence.

## Step 4: Produce Output

Structure your review as follows. Include file paths and line numbers for every finding.

### Critical (must fix)
Issues that would cause bugs, security vulnerabilities, data loss, panics, or incorrect behavior.
If none, write "None found."

### Important (should fix)
Issues affecting correctness, maintainability, or deviating from established project conventions.
If none, write "None found."

### Suggestions (consider)
Style improvements, Go 1.26 modernization opportunities, minor optimizations, DRY improvements.
If none, write "None found."

### Positive (good patterns)
Notable good patterns worth calling out. Always include at least one — reinforcing good practice matters.

### Summary
A 2-3 sentence overall assessment of the change quality and any themes across findings.
