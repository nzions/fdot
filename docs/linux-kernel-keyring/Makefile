# Makefile for Linux Kernel Keyring Documentation
# Demonstrates the kernel keyring "possession" permission model

CC := gcc
CFLAGS := -Wall -Wextra -std=c11

.PHONY: all clean help run demo

all: test_possession

# Build the demonstration program
test_possession: test_possession.c
	@echo "Building test_possession..."
	@$(CC) $(CFLAGS) -o $@ $<
	@echo "✅ Build complete"

# Help target
help:
	@echo "Linux Kernel Keyring Demonstration"
	@echo "==================================="
	@echo ""
	@echo "This program demonstrates the Linux kernel keyring 'possession'"
	@echo "permission model and why linking @u into @s is necessary."
	@echo ""
	@echo "Targets:"
	@echo "  all    - Build the demonstration program (default)"
	@echo "  run    - Build and run the demonstration"
	@echo "  demo   - Alias for 'run'"
	@echo "  clean  - Remove compiled binary"
	@echo "  help   - Show this help message"
	@echo ""
	@echo "Usage:"
	@echo "  make           # Build"
	@echo "  make run       # Build and run"
	@echo "  ./test_possession  # Run directly"
	@echo ""
	@echo "See README.md for detailed documentation"

# Run the demonstration
run: test_possession
	@echo ""
	@echo "Running Kernel Keyring Possession Demonstration..."
	@echo "=================================================="
	@echo ""
	@./test_possession

# Alias for run
demo: run

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -f test_possession
	@echo "✅ Clean complete"
