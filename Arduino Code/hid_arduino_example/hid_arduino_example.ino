#include "HID-Project.h"

// Extract parts of the __DATE__ and __TIME__ macros to build a unique ID
#define YEAR   ((__DATE__[7]-'0')*1000 + (__DATE__[8]-'0')*100 + (__DATE__[9]-'0')*10 + (__DATE__[10]-'0'))
#define MONTH  ((__DATE__[0]=='J' && __DATE__[1]=='a') ? 1 : \
                (__DATE__[0]=='F') ? 2 : \
                (__DATE__[0]=='M' && __DATE__[2]=='r') ? 3 : \
                (__DATE__[0]=='A' && __DATE__[1]=='p') ? 4 : \
                (__DATE__[0]=='M' && __DATE__[2]=='y') ? 5 : \
                (__DATE__[0]=='J' && __DATE__[2]=='n') ? 6 : \
                (__DATE__[0]=='J' && __DATE__[2]=='l') ? 7 : \
                (__DATE__[0]=='A' && __DATE__[1]=='u') ? 8 : \
                (__DATE__[0]=='S') ? 9 : \
                (__DATE__[0]=='O') ? 10 : \
                (__DATE__[0]=='N') ? 11 : 12)

#define DAY    ((__DATE__[4]==' ' ? 0 : (__DATE__[4]-'0'))*10 + (__DATE__[5]-'0'))
#define HOUR   ((__TIME__[0]-'0')*10 + (__TIME__[1]-'0'))
#define MINUTE ((__TIME__[3]-'0')*10 + (__TIME__[4]-'0'))
#define SECOND ((__TIME__[6]-'0')*10 + (__TIME__[7]-'0'))

// Build a "serial number" style ID: YYYYMMDDHHMMSS
#define SERIAL_NUM ((unsigned long long)YEAR*10000000000ULL + \
                    (unsigned long long)MONTH*100000000ULL + \
                    (unsigned long long)DAY*1000000ULL + \
                    (unsigned long long)HOUR*10000ULL + \
                    (unsigned long long)MINUTE*100ULL + \
                    (unsigned long long)SECOND)

const unsigned long deviceSerial = SERIAL_NUM;



#define PACKET_SIZE 64
uint8_t rawhidData[PACKET_SIZE];

String jsonData;



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
  Serial.print("Device ID: ");
  Serial.println(deviceSerial);

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
