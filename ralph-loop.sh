#!/bin/bash
# ralph-loop.sh — Ralph Wiggum loop for snip
# Usage: ./ralph-loop.sh [max_iterations]

set -e

MAX_ITERATIONS=${1:-20}
COUNT=0
PROMPT_FILE="PROMPT.md"
PROGRESS_FILE="progress.txt"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}  🔁 Ralph Wiggum Loop — snip${NC}"
echo -e "${CYAN}  Max iterations: ${MAX_ITERATIONS}${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Check prerequisites
if ! command -v claude &> /dev/null; then
    # Try common alternatives
    if command -v opencode &> /dev/null; then
        AGENT_CMD="opencode"
    else
        echo -e "${RED}Error: No AI coding agent found (claude, opencode).${NC}"
        echo "Install Claude Code: npm install -g @anthropic-ai/claude-code"
        exit 1
    fi
else
    AGENT_CMD="claude"
fi

if [ ! -f "$PROMPT_FILE" ]; then
    echo -e "${RED}Error: $PROMPT_FILE not found.${NC}"
    exit 1
fi

# Initialize progress file if missing
if [ ! -f "$PROGRESS_FILE" ]; then
    echo "# snip — Ralph Loop Progress Log" > "$PROGRESS_FILE"
    echo "" >> "$PROGRESS_FILE"
fi

# Initialize git if needed
if [ ! -d ".git" ]; then
    echo -e "${YELLOW}Initializing git repo...${NC}"
    git init
    git add -A
    git commit -m "Initial Ralph loop setup"
fi

while [ $COUNT -lt $MAX_ITERATIONS ]; do
    COUNT=$((COUNT + 1))
    TIMESTAMP=$(date '+%H:%M:%S')

    echo ""
    echo -e "${GREEN}━━━ Iteration ${COUNT}/${MAX_ITERATIONS} ━━━ ${TIMESTAMP} ━━━${NC}"

    # Show current task
    NEXT_TASK=$(python3 -c "
import json
with open('prd.json') as f:
    data = json.load(f)
for t in data['tasks']:
    if not t['done']:
        print(f\"Task {t['id']}: {t['title']}\")
        break
else:
    print('ALL TASKS COMPLETE')
" 2>/dev/null || echo "Could not read prd.json")

    echo -e "${YELLOW}  → ${NEXT_TASK}${NC}"

    if [ "$NEXT_TASK" = "ALL TASKS COMPLETE" ]; then
        echo -e "${GREEN}All tasks marked done. Exiting.${NC}"
        exit 0
    fi

    # Run the agent with the prompt
    # Capture output to check for completion signal
    OUTPUT_FILE=$(mktemp)
    echo -e "${CYAN}  ⏳ Running ${AGENT_CMD} (may take 1–2 min before you see output)...${NC}"
    echo ""

    if [ "$AGENT_CMD" = "claude" ]; then
        cat "$PROMPT_FILE" | claude -p --output-format text 2>&1 | tee "$OUTPUT_FILE"
    else
        cat "$PROMPT_FILE" | $AGENT_CMD 2>&1 | tee "$OUTPUT_FILE"
    fi

    echo ""
    echo -e "${CYAN}  ✓ Agent finished iteration ${COUNT}.${NC}"

    # Check for completion signal
    if grep -q "<promise>COMPLETE</promise>" "$OUTPUT_FILE" 2>/dev/null; then
        echo ""
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}  ✅ Ralph says: COMPLETE${NC}"
        echo -e "${GREEN}  Finished in ${COUNT} iteration(s)${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        rm -f "$OUTPUT_FILE"
        exit 0
    fi

    rm -f "$OUTPUT_FILE"

    # Brief pause between iterations to avoid rate limits
    echo -e "${CYAN}  Pausing 5s before next iteration...${NC}"
    sleep 5
done

echo ""
echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${RED}  ⚠️  Max iterations (${MAX_ITERATIONS}) reached.${NC}"
echo -e "${RED}  Check progress.txt and prd.json for status.${NC}"
echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
exit 1
