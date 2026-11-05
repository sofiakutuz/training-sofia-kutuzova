#!/bin/sh
# Spuštění služby objednávky. Sem časem přibude příprava (migrace, kontrola nastavení).
echo "startuji objednávky, adresář dat: /data"
exec /usr/local/bin/objednavky -listen :8080
