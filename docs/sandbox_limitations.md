# Python Sandbox Limitations & Security Disclosure

## Implementation Strategy
The QuantAlpha Python sandbox uses a **Denylist + Restricted Builtins** approach. While effective for common accidents and basic malicious scripts, it has inherent limitations compared to OS-level isolation.

## Known Limitations

### 1. Denylist Bypass Risk
Python's highly dynamic nature means there are often multiple ways to reach a specific function or attribute. While we block `__class__`, `__mro__`, and `getattr`, an attacker might find other paths through standard library modules or third-party packages (like `numpy`) that have not been audited.

### 2. Resource Exhaustion (CPU)
The sandbox uses a `time.perf_counter()` check in `run_backtest`. However, if the user script performs a heavy computation *inside* a single `signal(row)` call (e.g., a massive matrix inversion or an infinite loop without yielding), the worker thread will hang until that call returns or the process is killed.

### 3. Resource Exhaustion (Memory)
There are currently no strict per-process memory limits (cgroups/ulimits) enforced within the Python code. A script can allocate large arrays in `numpy` until it triggers an Out-Of-Memory (OOM) killer on the host or container.

### 4. Side-Channel Attacks
The sandbox does not protect against timing attacks or other side-channel leaks of system information.

## Security Recommendations
1. **Never use for Public Multi-Tenant**: This sandbox is designed for internal quant researchers who are trusted but may write buggy or accidentally destructive code. It is **NOT** secure enough for a public platform where arbitrary users can submit code.
2. **Container Isolation**: Always run the worker in a container with strict CPU and memory limits (`deploy.resources.limits`).
3. **Audit**: Regularly review user-submitted alpha scripts for suspicious patterns.
4. **Future Path**: For higher security, migrate to a WebAssembly (WASM) runner or a micro-VM (like Firecracker) for script execution.
