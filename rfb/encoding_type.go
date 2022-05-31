package rfb

type EncodingType int32

//go:generate stringer -type=EncodingType

// https://www.iana.org/assignments/rfb/rfb.xml#rfb-4
const (
	EncRaw                           EncodingType = 0 // 不编码，原始的格式
	EncCopyRect                      EncodingType = 1 //从帧缓冲复制
	EncRRE                           EncodingType = 2 // 二维游程编码
	EncCoRRE                         EncodingType = 4 // 二维游程编码的变体
	EncHexTile                       EncodingType = 5 // RRE 的变种，图块游程编码
	EncZlib                          EncodingType = 6 // zlib压缩
	EncTight                         EncodingType = 7 // tightvnc项目设置的编码
	EncZlibHex                       EncodingType = 8 // zlib压缩Hextile
	EncUltra1                        EncodingType = 9
	EncUltra2                        EncodingType = 10
	EncTRLE                          EncodingType = 15 //图块游程编码
	EncZRLE                          EncodingType = 16 //zlib 压缩的游程编码
	EncJPEG                          EncodingType = 21
	EncJRLE                          EncodingType = 22
	EncAtenAST2100                   EncodingType = 87
	EncAtenASTJPEG                   EncodingType = 88
	EncAtenHermon                    EncodingType = 89
	EncAtenYarkon                    EncodingType = 90
	EncAtenPilot3                    EncodingType = 91
	EncJPEGQualityLevelPseudo10      EncodingType = -23
	EncJPEGQualityLevelPseudo9       EncodingType = -24
	EncJPEGQualityLevelPseudo8       EncodingType = -25
	EncJPEGQualityLevelPseudo7       EncodingType = -26
	EncJPEGQualityLevelPseudo6       EncodingType = -27
	EncJPEGQualityLevelPseudo5       EncodingType = -28
	EncJPEGQualityLevelPseudo4       EncodingType = -29
	EncJPEGQualityLevelPseudo3       EncodingType = -30
	EncJPEGQualityLevelPseudo2       EncodingType = -31
	EncJPEGQualityLevelPseudo1       EncodingType = -32
	EncDesktopSizePseudo             EncodingType = -223 //桌面分辨率伪编码
	EncLastRectPseudo                EncodingType = -224 //  表示是最后一个矩形的伪编码
	EncPointerPosPseudo              EncodingType = -232
	EncCursorPseudo                  EncodingType = -239 //光标掩码
	EncXCursorPseudo                 EncodingType = -240
	EncCompressionLevel10            EncodingType = -247
	EncCompressionLevel9             EncodingType = -248
	EncCompressionLevel8             EncodingType = -249
	EncCompressionLevel7             EncodingType = -250
	EncCompressionLevel6             EncodingType = -251
	EncCompressionLevel5             EncodingType = -252
	EncCompressionLevel4             EncodingType = -253
	EncCompressionLevel3             EncodingType = -254
	EncCompressionLevel2             EncodingType = -255
	EncCompressionLevel1             EncodingType = -256
	EncQEMUPointerMotionChangePseudo EncodingType = -257
	EncQEMUExtendedKeyEventPseudo    EncodingType = -258
	EncTightPng                      EncodingType = -260
	EncLedStatePseudo                EncodingType = -261
	EncDesktopNamePseudo             EncodingType = -307
	EncExtendedDesktopSizePseudo     EncodingType = -308
	EncXvpPseudo                     EncodingType = -309
	EncClientRedirect                EncodingType = -311
	EncFencePseudo                   EncodingType = -312
	EncContinuousUpdatesPseudo       EncodingType = -313
	EncCursorWithAlphaPseudo         EncodingType = -314
	EncExtendedClipboardPseudo       EncodingType = -1063131698 //C0A1E5CE
	EncTightPNGBase64                EncodingType = 21 + 0x574d5600
	EncTightDiffComp                 EncodingType = 22 + 0x574d5600
	EncVMWDefineCursor               EncodingType = 100 + 0x574d5600
	EncVMWCursorState                EncodingType = 101 + 0x574d5600
	EncVMWCursorPosition             EncodingType = 102 + 0x574d5600
	EncVMWTypematicInfo              EncodingType = 103 + 0x574d5600
	EncVMWLEDState                   EncodingType = 104 + 0x574d5600
	EncVMWServerPush2                EncodingType = 123 + 0x574d5600
	EncVMWServerCaps                 EncodingType = 122 + 0x574d5600
	EncVMWFrameStamp                 EncodingType = 124 + 0x574d5600
	EncOffscreenCopyRect             EncodingType = 126 + 0x574d5600
)
