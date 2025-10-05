#include "HID-Project.h"


#define PACKET_SIZE 64
uint8_t rawhidData[PACKET_SIZE];

String jsonData;


// Serial Number Generator
unsigned long makeSerial() {
  // Example: 2025-10-04 15:32:10 → 5100432
  const char* date = __DATE__;  // "Oct  4 2025"
  const char* time = __TIME__;  // "15:32:10"

  // Convert parts
  int month =
    (date[0] == 'J' && date[1] == 'a') ? 1 : (date[0] == 'F')                   ? 2
                                           : (date[0] == 'M' && date[2] == 'r') ? 3
                                           : (date[0] == 'A' && date[1] == 'p') ? 4
                                           : (date[0] == 'M' && date[2] == 'y') ? 5
                                           : (date[0] == 'J' && date[2] == 'n') ? 6
                                           : (date[0] == 'J' && date[2] == 'l') ? 7
                                           : (date[0] == 'A' && date[1] == 'u') ? 8
                                           : (date[0] == 'S')                   ? 9
                                           : (date[0] == 'O')                   ? 10
                                           : (date[0] == 'N')                   ? 11
                                                                                : 12;

  int day = ((date[4] == ' ') ? 0 : (date[4] - '0')) * 10 + (date[5] - '0');
  int hour = (time[0] - '0') * 10 + (time[1] - '0');
  int minute = (time[3] - '0') * 10 + (time[4] - '0');
  int second = (time[6] - '0') * 10 + (time[7] - '0');

  // Combine into a compact unique number (month*1000000 + day*10000 + hour*100 + minute)
  return (month * 1000000UL) + (day * 10000UL) + (hour * 100UL) + minute;
}


const unsigned long deviceSerial = makeSerial();


// --- Analog Sliders ---
const int NUM_SLIDERS = 2;
const int analogInputs[NUM_SLIDERS] = { 10, 18 };
int analogSliderValues[NUM_SLIDERS];

// --- Buttons ---
const int NUM_BUTTONS = 3;
const int buttonPins[NUM_BUTTONS] = { 2, 3, 4 };
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
    }

    if (rxBuffer[0] == 5) {
      delay(10);
      sendData();
    } else if (rxBuffer[0] == 6) {
      delay(20);
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
  data = "{\"id\":" + String(deviceSerial) + ", ";
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
  Serial.println(data);
  return data;
}

