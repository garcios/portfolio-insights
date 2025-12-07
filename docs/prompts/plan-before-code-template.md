# Prompts for "Plan Before Code" Strategy

Implementing a "Plan Before Code" step is effective for leveraging AI in complex engineering tasks. Below are three increasingly detailed prompts to help structure the work.

## 1. Simple Task: File Modification (Focus on Impact)
Use this for small, well-defined tasks like fixing a specific bug or updating a function signature.

### Prompt
> "I need you to [Describe Task e.g., fix a bug where...]. The relevant file is [File Path], specifically the function [Function Name].
>
> **Before writing any code**, provide a **Change Plan** detailing:
>
> *   **Diagnosis**: What is the probable cause of the bug?
> *   **Affected Modules**: List the files/functions that will be directly touched by the fix.
> *   **Code Strategy**: Briefly describe the logical change you will implement (e.g., changing from an inclusive to an exclusive range, adding a filter).
> *   **Test Case**: Provide the specific unit test assertion (input and expected output) you will create to confirm the fix."

---

## 2. Moderate Task: Feature Implementation (Focus on Steps)
Use this for new features that involve changes across multiple files or layers (e.g., adding a new API endpoint, creating a new component).

### Prompt
> "Implement a new feature to [Describe Feature e.g., allow users to archive old projects]. This will involve changes to [List Areas e.g., API, database, frontend].
>
> **Before writing any code**, create a **Detailed Strategy Document** with the following sections:
>
> ### I. High-Level Overview
> A brief summary of the feature and its core goal.
>
> ### II. Implementation Steps (Decomposition)
> Break the task down into sequential, atomic steps (e.g., 1. Add column... 2. Create endpoint... 3. Implement frontend...).
>
> ### III. Affected Files & Context
> For each step, list the files that will be modified and briefly explain why."

---

## 3. Complex Task: Refactoring/Architecture Change (Focus on Risk & Alternatives)
Use this when the task involves non-trivial restructuring, migration, or changes to core architectural components.

### Prompt
> "Refactor the [System Component] from [Old State] to [New State]. This is a high-risk change impacting [Impact Areas].
>
> **Before implementing anything**, provide a **comprehensive Design Proposal** that includes:
>
> ### 1. Architectural Impact & Risk Assessment
> *   Briefly describe the core components that will be removed, added, or modified.
> *   Identify the top three risks associated with this change (e.g., security flaws, session invalidation, downtime).
>
> ### 2. Proposed Implementation Strategy (Phased Approach)
> *   Outline the change in phases. How can we deploy this without breaking current user sessions?
>
> ### 3. Alternatives Considered
> *   What is one alternative approach and why is the chosen one preferred?
>
> ### 4. Go/No-Go Checklist
> *   List the key milestones that must be achieved before merging the final code (e.g., All legacy tests pass, new security tests added, rollback plan defined)."
