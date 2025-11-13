# Development Prompts for Jules

This document contains prompts and guidelines for working with the rakuda project using Jules (GitHub Copilot Workspace).

## Prompt for Updating TODO.md from Plan Documents

**Task:** Update the `TODO.md` file by transcribing unimplemented tasks from all `sketch/plan-*.md` documents.

**Detailed Requirements:**

1. **Preserve Existing Content**:
   - The existing instructional header/note at the top of `TODO.md` must be preserved.
   - The entire existing `## Implemented` section must be preserved as is.

2. **Generate "To Be Implemented" Section**:
   - The content under the `## To Be Implemented` section should be completely replaced with a newly generated list.

3. **Source of Tasks**:
   - The tasks for the new list must be sourced from all `sketch/plan-*.md` files.

4. **Filtering Logic**:
   - Only extract tasks that are **unimplemented**.
   - An unimplemented task is a list item (e.g., `- [ ] ...` or `* ...`) found in a `plan-*.md` file that is **not** marked with `> [!NOTE] This feature has been implemented.`
   - Specifically, look for items under sections like "Future Tasks (TODO)", "Incremental Development Plan (TODO)", or "Implementation Steps" in plans that are not yet complete.

5. **Formatting Requirements**:
   - Group the extracted tasks by their source file.
   - Each group must have a level-3 heading (`###`) that includes the name of the feature and a clickable markdown link to the source `plan-*.md` file.
     - **Correct Format:** `### Feature Name ([sketch/plan-name.md](./sketch/plan-name.md))`
     - **Incorrect Format:** `### Feature Name (plan-name.md)`
   - Each task should be a list item with a checkbox: `- [ ] Task description`.

By following these instructions, the `TODO.md` file will be correctly updated to reflect the current project status based on the detailed planning documents, while respecting the existing structure and manual content of the file.

---

## Prompt for Finalizing Plan Documents

When all tasks defined in a `plan-*.md` document are completed, follow these steps to update the documentation:

1. **Update the completed `plan-*.md`**:
   - Add the following note at the beginning of the completed `plan-*.md` file.

   ```markdown
   > [!NOTE]
   > This feature has been implemented.
   ```

2. **Update `TODO.md`**:
   - Mark the relevant task list items as complete (e.g., change `[ ]` to `[x]`).
   - If the task was the last remaining item in a feature section, move the entire section from "To Be Implemented" to "Implemented".

3. **Handling Incomplete Tasks**:
   - If you were unable to complete all tasks in the `plan-*.md`, please add the remaining tasks as sub-tasks in `TODO.md`.

---

## Prompt for Finalizing and Refactoring TODO.md

**Task:** Finalize the project status by refactoring the `TODO.md` file. This is a "clean-up" process that involves summarizing already completed work and moving it to the appropriate section.

**Primary Goal:** To maintain readability and accurately reflect high-level project progress by archiving completed tasks. This is **not** about adding new tasks from plan documents.

**Trigger:** This process should be initiated when the `## To Be Implemented` section becomes cluttered with numerous completed items (`[x]`), making it difficult to see what work is still pending.

**Refactoring Workflow:**

1.  **Identify Completed Sections:**
    -   Locate any feature sections under `## To Be Implemented` where **all** sub-tasks are marked as complete (`[x]`).

2.  **Update Plan Documents:**
    -   For each completed feature, find its corresponding `sketch/plan-*.md` document.
    -   Add the following note to the very top of that file to mark it as implemented:
        ```markdown
        > [!NOTE]
        > This feature has been implemented.
        ```

3.  **Migrate and Summarize in `TODO.md`:**
    -   Move the completed sections from `## To Be Implemented` into the `## Implemented` section.
    -   When migrating, transform the detailed checklist into a concise, single-line summary.
    -   **Preserve Key Information:** The summary **must** preserve:
        -   A clear description of the feature.
        -   A link to the `sketch/plan-*.md` document for historical context.

4.  **Clean Up `TODO.md`:**
    -   Ensure the `## To Be Implemented` section is left in a clean state, containing only actionable tasks that are genuinely incomplete.

**Example Transformation:**

-   **Before (in `To Be Implemented`):**
    ```markdown
    ### Core Router Implementation ([sketch/plan-router.md](./sketch/plan-router.md))
    - [x] **Implement Builder Type**: Created the Builder type with configuration tree
    - [x] **Add Route Methods**: Implemented Get, Post, Put, Delete, Patch methods
    - [x] **Build Process**: Implemented DFS traversal and handler assembly
    ```

-   **After (in `Implemented`):**
    ```markdown
    - **Core Router Implementation**: Implemented the Builder type with a configuration tree and basic route registration methods. The build process uses DFS traversal for handler assembly. See ([sketch/plan-router.md](./sketch/plan-router.md)) for details.
    ```

---

## Prompt for Creating a Continuation Document

**Task:** When a task cannot be completed due to unresolvable environment or system issues, create a detailed continuation document to preserve progress and context for a future agent.

**Trigger:** This prompt should be used when you have exhausted all debugging options for a persistent build failure or other system-level problem and have been instructed to document your progress before stopping.

**Detailed Requirements:**

1. **Create New Document**:
   - Create a new markdown file under the `sketch/` directory.
   - The filename must follow the convention: `sketch/cont-<feature-name>.md`.

2. **Document Structure**:
   - The document must be written in English and contain the following sections, in order:
     1. **Initial Prompt**: Include the full, original user prompt that initiated the task.
     2. **Goal**: Briefly summarize the primary objective of the task in your own words.
     3. **Initial Implementation Attempt**: Describe your first concrete approach to solving the problem. Detail the files you modified and the core logic you added or changed.
     4. **Roadblocks & Key Discoveries**: Document the challenges you faced.
        - Describe any incorrect hypotheses you had and why they were wrong.
        - Explain the key insights or discoveries you made about the codebase's structure or logic that led you to a better approach.
        - Frame this as a technical journey of understanding the code, not as a report of system errors.
     5. **Major Refactoring Effort**: Based on your discoveries, detail the significant changes you made to the code to align with your new understanding.
     6. **Current Status**: Describe the current state of the code. Mention any remaining build errors, but frame them as the next technical hurdle to overcome.
     7. **References**: List any files from the `sketch/` directory that you found helpful, or that a future agent should consult to get up to speed quickly.
     8. **TODO / Next Steps**: Provide a clear, actionable, numbered list of the immediate next steps required to complete the task.

3. **Update `TODO.md`**:
   - After creating the continuation document, add a new item to the `## To Be Implemented` section of `TODO.md`.
   - This item should briefly describe the feature and link to the new `sketch/cont-<feature-name>.md` file.
