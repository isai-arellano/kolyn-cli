#!/bin/bash
set -e

# Configuration
BINARY="kolyn"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}Uninstalling Kolyn CLI...${NC}\n"

# Check if kolyn is installed
if [ ! -f "$INSTALL_DIR/$BINARY" ]; then
    echo -e "${YELLOW}Kolyn is not installed in $INSTALL_DIR${NC}"
    echo -e "Checking other common locations...\n"
    
    # Check in user's bin
    if [ -f "$HOME/bin/$BINARY" ]; then
        INSTALL_DIR="$HOME/bin"
        echo -e "Found in: ${GREEN}$INSTALL_DIR${NC}"
    elif [ -f "$HOME/.local/bin/$BINARY" ]; then
        INSTALL_DIR="$HOME/.local/bin"
        echo -e "Found in: ${GREEN}$INSTALL_DIR${NC}"
    else
        echo -e "${RED}Kolyn binary not found.${NC}"
        exit 1
    fi
fi

# Remove binary
echo -e "Removing binary from $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    rm -f "$INSTALL_DIR/$BINARY"
else
    sudo rm -f "$INSTALL_DIR/$BINARY"
fi

echo -e "${GREEN}Kolyn binary removed successfully!${NC}\n"

# Ask about Docker services
echo -e "\n${YELLOW}Do you want to remove Docker services created by Kolyn? [y/N]${NC}"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    if [ -d "$HOME/.kolyn/services" ]; then
        echo -e "Removing ~/.kolyn/services directory..."
        
        # Stop all services first
        echo -e "${BLUE}Stopping all services...${NC}"
        for dir in "$HOME/.kolyn/services"/*; do
            if [ -d "$dir" ] && [ -f "$dir/docker-compose.yml" ]; then
                echo -e "  Stopping $(basename "$dir")..."
                (cd "$dir" && docker compose down -v 2>/dev/null || true)
            fi
        done
        
        rm -rf "$HOME/.kolyn/services"
        echo -e "${GREEN}Docker services removed!${NC}"
    else
        echo -e "${BLUE}No Docker services found.${NC}"
    fi
fi

# Ask about Skills preservation
echo -e "\n${YELLOW}Do you want to remove configuration files? [y/N]${NC}"
read -r response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    if [ -d "$HOME/.kolyn" ]; then
        echo -e "${YELLOW}Do you want to KEEP your downloaded skills/sources? (Recommended if you plan to reinstall) [Y/n]${NC}"
        read -r keep_skills
        
        if [[ "$keep_skills" =~ ^([nN][oO]|[nN])$ ]]; then
            echo -e "Removing entire ~/.kolyn directory..."
            rm -rf "$HOME/.kolyn"
            echo -e "${GREEN}All configuration and skills removed!${NC}"
        else
            echo -e "Cleaning up configuration but keeping skills..."
            # Remove everything except skills and sources
            find "$HOME/.kolyn" -mindepth 1 -maxdepth 1 ! -name 'skills' ! -name 'sources' -exec rm -rf {} +
            echo -e "${GREEN}Configuration removed. Skills preserved in ~/.kolyn/skills${NC}"
        fi
    else
        echo -e "${BLUE}No configuration files found.${NC}"
    fi
fi

echo -e "\n${GREEN}✓ Kolyn has been uninstalled successfully!${NC}"
echo -e "Thank you for using Kolyn CLI.\n"
