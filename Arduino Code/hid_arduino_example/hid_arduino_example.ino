#include "HID-Project.h"

#define PACKET_SIZE 64
uint8_t rawhidData[PACKET_SIZE];

String jsonData;



// --- Analog Sliders ---
const int NUM_SLIDERS = 1;
const int analogInputs[NUM_SLIDERS] = { 10 };
int analogSliderValues[NUM_SLIDERS];

// --- Buttons ---
const int NUM_BUTTONS = 1;
const int buttonPins[NUM_BUTTONS] = { 2 };
int buttonStates[NUM_BUTTONS];
bool buttonTriggered[NUM_BUTTONS];


void setup() {
  // Setup Serial
  Serial.begin(115200);
  // Initiate Analog Sliders
  for (int i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }

  // Initiate Buttions
  for (int i = 0; i < NUM_BUTTONS; i++) {
    pinMode(buttonPins[i], INPUT_PULLUP);
  }

  // Start RawHID with buffer + size
  RawHID.begin(rawhidData, sizeof(rawhidData));
}

void loop() {
  // -------- Receive packet from host --------
  int bytesAvailable = RawHID.available();
  uint8_t rxBuffer[PACKET_SIZE];


  if (bytesAvailable > 0) {
    // loop through all the bits in the buffer
    for (int i = 0; i < bytesAvailable; i++) {
      byte b = RawHID.read();  // read one byte at a time
      if (b >= 0) {
        rxBuffer[i] = b;
      }

      if (rxBuffer[0] == 5) {
        delay(10);
        sendData();
      }
      memset(rxBuffer, 0, sizeof(rxBuffer));  // clear rxBuffer
      b = 0;                                  // reset byte count
    }
  }
}


void sendData() {
  updateSliderValues();
  updateButtons();
  jsonData = createJsonPacket();


  // -------- Send raw hid packet --------
  for (int x = 0; x < jsonData.length(); x = x) {
    uint8_t txBuffer[PACKET_SIZE];
    for (uint8_t i = 0; i < PACKET_SIZE; i++) {
      txBuffer[i] = jsonData[x];
      x++;
    }
    RawHID.write(txBuffer, sizeof(txBuffer));
  }
}


void updateSliderValues() {
  for (int i = 0; i < NUM_SLIDERS; i++) {
    analogSliderValues[i] = analogRead(analogInputs[i]);
  }
}

void updateButtons() {
  for (int i = 0; i < NUM_BUTTONS; i++) {
    bool button = digitalRead(buttonPins[i]);

    // invert because of INPUT_PULLUP (pressed = 0)
    if (!button && !buttonTriggered[i]) {
      buttonStates[i] = 1;
      buttonTriggered[i] = true;
    } else if (button) {
      buttonTriggered[i] = false;
      buttonStates[i] = 0;
    } else {
      buttonStates[i] = 0;
    }
  }
}

String createJsonPacket() {
  String data;
  data = "{\"s\":[";
  for (int i = 0; i < NUM_SLIDERS; i++) {
    data += analogSliderValues[i] / 4;  // scale 0–1023 → 0–255
    if (i < NUM_SLIDERS - 1) {
      data += ",";
    }
  }
  data += "], \"b\":[";
  for (int i = 0; i < NUM_BUTTONS; i++) {
    data += buttonStates[i];
    if (i < NUM_BUTTONS - 1) {
      data += ",";
    }
  }
  data += "]}";
  return data;
}
