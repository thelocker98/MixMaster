#include "HID-Project.h"


#define PACKET_SIZE 64
uint8_t rawhidData[PACKET_SIZE];

String jsonData;


const String deviceSerial = "H0S05B10T0001";

// --- Analog Sliders ---
const int NUM_SLIDERS = 5;
const int analogInputs[NUM_SLIDERS] = { A0, A1, A2, A3, A6 };
int analogSliderValues[NUM_SLIDERS];

// --- Buttons ---
const int NUM_BUTTONS = 10;
const int buttonPins[NUM_BUTTONS] = { 5, 6, 7, 8, 9, 10, 15, 16, 14, 3 };
int buttonStates[NUM_BUTTONS];
bool buttonTriggered[NUM_BUTTONS];



void setup() {
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
    }

    if (rxBuffer[0] == 'D') {
      delay(1);
      sendData();
    } else if (rxBuffer[0] == 'I') {
      delay(1);
      sendSerialNum(String(deviceSerial));
    }
    memset(rxBuffer, 0, sizeof(rxBuffer));  // clear rxBuffer                               // reset byte count
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

void sendSerialNum(String serialNum) {
  serialNum = "{\"Id\":\"" + serialNum + "\"}";

  // -------- Send raw hid packet --------
  for (int x = 0; x < serialNum.length(); x = x) {
    uint8_t txBuffer[PACKET_SIZE];
    for (uint8_t i = 0; i < PACKET_SIZE; i++) {
      txBuffer[i] = serialNum[x];
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
  data = "{\"id\":\"" + String(deviceSerial) + "\", ";
  data += "\"s\":[";
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
