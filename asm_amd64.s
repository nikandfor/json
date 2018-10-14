
#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

TEXT	·skipSpaces(SB), NOSPLIT, $0-40
	MOVQ	s+32(SP), BX
	MOVQ	blen+16(SP), CX
	MOVQ	bbuf+8(SP), DX
	JMP	first
loop:
	INCQ	BX
first:
	CMPQ	BX, CX
	JGE	retend
	MOVBLZX	(DX)(BX*1), SI
	CMPB	SIB, $32
	JEQ	loop
	LEAL	-9(SI), DI
	CMPB	DIB, $1
	JLS	loop
retend:
	MOVQ	BX, r2+40(SP)
	RET
