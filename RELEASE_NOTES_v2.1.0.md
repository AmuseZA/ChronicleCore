# ChronicleCore v2.1.0

This release introduces **Universal Deep Tracking**, a major upgrade to the activity detection engine that provides richer context for all your applications, including those that were previously hard to track (like Opera Sidebars).

## 🚀 New Features

### Universal Deep Tracking
ChronicleCore now uses advanced **UI Automation (UIA)** to "read" the screen just like a screen reader. This allows it to:
-   **See "Hidden" Windows**: Detects activity in apps that hide their window title from the standard OS API (e.g., Opera GX Sidebars).
-   **Extract Rich Context**: Instead of just "Word", see "Editing Proposal.docx". Instead of "Teams", see "Chatting with Startups".
-   **Browser Address Bar Detection**: Captures the exact URL from Chrome, Edge, and Opera even if the window title is generic.

### "God Mode" Sidebar Tracking
-   Fixed an issue where **WhatsApp** and other sidebars in Opera/Opera GX were ignored or marked as Idle.
-   The tracker now accurately bills time spent in these sidebars to the correct client/project.

## 🛠 Improvements
-   **Async Enrichment**: The deep tracking engine runs asynchronously in the background, ensuring **zero lag** or performance impact on your PC while switching windows.
-   **Smart Context**: Automatically prioritizes the most detailed information (Window Title vs. UIA Document Name).

## 📦 Upgrade Instructions

1.  **Download** the new installer below.
2.  **Run** `ChronicleCore_Setup_v2.1.0.exe`.
3.  The installer will automatically update your existing installation. Database and settings are preserved.
