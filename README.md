# phenomena-dna-scroll-intro
# Phenomena DNA Scroll Intro

Port Go/Ebitengine de l’intro Phenomena, disponible sur ordinateur et Android.

## Ordinateur

```sh
go run ./cmd/phenomena
```

Les flèches haut/bas règlent le volume. Un clic lance la séquence de sortie.

## Vérifications

```sh
go test ./...
go test -race ./...
go vet ./...
```

## Android

La configuration validée utilise Java 17, Android SDK 36, NDK r28.2,
Gradle 8.11.1, Ebitengine 2.9.11 et une cible `arm64-v8a`.

Avec un appareil Android autorisé en USB :

```sh
./scripts/run-android.sh
```

Le script génère l’AAR Ebitengine, compile l’APK de débogage, l’installe puis
lance `com.olivierh.phenomenadna/.MainActivity`. Un toucher pendant la scène
principale lance la séquence de sortie, comme le clic sur ordinateur.
