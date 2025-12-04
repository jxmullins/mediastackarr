# MediaStack iOS App Plan

## Objective
Create a native iOS application to manage the MediaStack, replicating the functionality of the `mediastack` CLI tool. The app will allow users to monitor, deploy, and manage their media server remotely from an iPhone.

## Architecture

### Core Technology
*   **Platform**: iOS 16.0+
*   **Language**: Swift 5+
*   **UI Framework**: SwiftUI
*   **Concurrency**: Swift Async/Await

### Communication Layer
The app will communicate with the MediaStack server via **SSH**. This approach allows the app to execute the existing `mediastack` CLI commands directly on the server without requiring a separate API backend to be installed.

*   **Library**: `Citadel` or `Shout` (Swift SSH wrappers).
*   **Authentication**: SSH Key (PEM/OpenSSH) or Password.

## Features & Scope

### 1. Connection Management
*   Save multiple server profiles (Host, Port, User, Auth).
*   Test connection status.
*   Secure storage of credentials (Keychain).

### 2. Dashboard (Home)
*   **System Status**: CPU/Memory usage of the host (via `htop` or `docker stats`).
*   **Stack Status**: Quick summary of running/stopped containers.
*   **Quick Actions**: "Restart All", "Stop All".

### 3. Services Management
*   **List View**: detailed list of all containers (derived from `mediastack status --json`).
    *   Status indicators (Running, Healthy, Stopped).
    *   Uptime.
    *   Image version.
*   **Detail View**:
    *   Start / Stop / Restart individual service.
    *   View Logs (`mediastack logs <service>`).

### 4. Deployment
*   **Deploy Interface**:
    *   Select variant (Full / Mini / No-VPN).
    *   Toggle options (Pull images, Prune).
    *   "Deploy" button with progress log streaming.

### 5. Configuration
*   **View Config**: Read-only view of `.env` file.
*   **API Keys**: Display extracted API keys (`mediastack apikeys --json`).

## Technical Implementation Steps

### Phase 1: Foundation
1.  Initialize Xcode project.
2.  Implement SSH connection manager using `Citadel`.
3.  Create `CommandRunner` service to execute shell commands remotely.

### Phase 2: Core Data
1.  Implement `StatusParser` to parse JSON output from `mediastack status --json`.
2.  Build the `ServicesListViewModel` to fetch and store container state.

### Phase 3: UI Construction
1.  **DashboardView**: Summary cards.
2.  **ServiceListView**: List with Swipe-to-action (Restart/Stop).
3.  **LogView**: Scrollable text view with ansi-color support.

### Phase 4: deployment Integration
1.  Implement streaming command execution for `deploy` (needs real-time output).
2.  Build `DeployView` with toggles for CLI flags.

## Dependencies
*   `Citadel` (SSH)
*   `KeychainAccess` (Credential storage)
*   `SwiftyJSON` or standard `Codable` (Parsing)

## Future Considerations
*   Push Notifications for service downtime (requires a background agent or polling).
*   File browser for `base-working-files`.
