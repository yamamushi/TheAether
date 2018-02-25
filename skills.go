package main

// GetSkillList function
func GetSkillList() (m map[string]string) {
	m = map[string]string{
		"-Acrobatics\n": "You can keep your balance while traversing narrow or treacherous surfaces. You can also dive, flip, jump, and roll, avoiding attacks and confusing your opponents.",
		"-Appraise\n": "A DC 20 Appraise check determines the value of a common item. If you succeed by 5 or more, you also determine " +
			"if the item has magic properties, although this success does not grant knowledge of the magic item’s abilities. If you fail the check by less than 5, " +
			"you determine the price of that item to within 20% of its actual value. ",
		"-Bluff\n": "You can use Bluff to pass hidden messages to another character without others understanding your true meaning. " +
			"The DC of this check is 15 for simple messages and 20 for complex messages. If you are successful, the target automatically understands you, " +
			"assuming you are speaking in a language that it understands. If your check fails by 5 or more, you deliver the wrong message. Other creatures " +
			"that hear the message can decipher the message by succeeding at an opposed Sense Motive check against your Bluff result.",
		"-Climb\n": "With a successful Climb check, you can advance up, down, or across a slope, wall, or other steep incline (or even across a ceiling, provided it has handholds) " +
			"at one-quarter your normal speed. A slope is considered to be any incline at an angle measuring less than 60 degrees; a wall is any incline at an angle measuring 60 degrees " +
			"or more. A Climb check that fails by 4 or less means that you make no progress, and one that fails by 5 or more means that you fall from whatever height you have already attained. ",
		//		"-Craft\n" +
		//		"-Diplomacy\n" +
		//		"-Disable-Device\n" +
		//		"-Disguise\n" +
		//		"-Excape-Artist\n" +
		//		"-Fly\n" +
		//		"-Handle-Animal\n" +
		//		"-Heal\n" +
		//		"-Intimidate\n" +
		//		"-Knowledge-(arcana)/(dungeoneering)/(engineering)/(geography)/(history)/(local)/(nature)/(nobility)(planes)/(religion)\n" +
		//		"-Linguistics\n" +
		//		"-Perception\n" +
		//		"-Perform\n" +
		//		"-Profession\n" +
		//		"-Ride\n"+
		//		"-Sense-Motive\n" +
		//		"-Sleight-of-Hand\n" +
		//		"-Spellcraft\n" +
		//		"-Stealth\n" +
		//		"-Survival\n" +
		//		"-Swim\n" +
		//		"-Use-Magic-Device\n"
	}

	return m

}
