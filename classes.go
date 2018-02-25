package main



func GetClassList() (m map[string]string) {

	m = map[string]string{
		"-Bard\n":			"-Untold wonders and secrets exist for those skillful enough to discover them. Through cleverness, talent, and magic, " +
							"these cunning few unravel the wiles of the world, becoming adept in the arts of persuasion, manipulation, and inspiration. ",

		"-Claric\n":		"-In faith and the miracles of the divine, many find a greater purpose. Called to serve powers beyond most mortal understanding, " +
							"all priests preach wonders and provide for the spiritual needs of their people. Clerics are more than mere priests, though; " +
							"these emissaries of the divine work the will of their deities through strength of arms and the magic of their gods. ",

		"-Druid\n":			"-Within the purity of the elements and the order of the wilds lingers a power beyond the marvels of civilization." +
							" Furtive yet undeniable, these primal magics are guarded over by servants of philosophical balance known as druids. " +
							"Allies to beasts and manipulators of nature, these often misunderstood protectors of the wild strive to shield their lands from " +
							"all who would threaten them and prove the might of the wilds to those who lock themselves behind city walls.",

		"-Elf\n":			"-Tall, noble, and often haughty, elves are long-lived and subtle masters of the wilderness. Elves excel in the arcane arts. Often they use their intrinsic link to nature to " +
							"forge new spells and create wondrous items that, like their creators, seem nearly impervious to the ravages of time. A private and often introverted race, " +
							"elves can give the impression they are indifferent to the plights of others.",

		"-Enchanter\n":		"-A devoted enchanter shines most brightly in more subtle settings, either in infiltration, " +
							"or simply gathering knowledge she shouldn't have. Where many mages have issues socially, it is a devoted enchanter's home field.",

		"-Fighter\n":		"-Some take up arms for glory, wealth, or revenge. Others do battle to prove themselves, to protect others, or because they know nothing else. " +
							"Still others learn the ways of weaponcraft to hone their bodies in battle and prove their mettle in the forge of war. Lords of the battlefield, " +
							"fighters are a disparate lot, training with many weapons or just one, perfecting the uses of armor, learning the fighting techniques of exotic masters, " +
							"and studying the art of combat, all to shape themselves into living weapons.",

		"-Monk\n":			"-For the truly exemplary, martial skill transcends the battlefield—it is a lifestyle, a doctrine, a state of mind. " +
							"These warrior-artists search out methods of battle beyond swords and shields, finding weapons within themselves just as capable of crippling " +
							"or killing as any blade. These monks (so called since they adhere to ancient philosophies and strict martial disciplines) elevate their bodies to " +
							"become weapons of war, from battle-minded ascetics to self-taught brawlers.",

		"-Necromancer\n":	"-While others use magic to do paltry things like conjure fire or fly, the Necromancer is a master over death itself. " +
							"They study the deep and forbidden secrets that raise the dead, controlling minions toward a variety of goals. Perhaps they seek " +
							"the power that mastery over death provides. Perhaps they are serious and unashamed scholars, who reject the small-minded boundaries held to by others. " +
							"Each enemy they fell becomes an eager and disposable ally, they become immune to the energies of death and decay, and ultimately harness the immortality and " +
							"power of undeath for themselves.",

		"-Ninja\n":			"-A ninja is one who refines stealth, intellegence gathering, powerful combat techniques, and mysticism into a " +
							"deadly science and sophisticated techniques of warfare. When the odds are unfavorable and dishonor threatens, the ninja can be hired to bring " +
							"victory and restore harmony of society through espionage and assassination.",

		"-Paladin\n":		"-Through a select, worthy few shines the power of the divine. Called paladins, " +
							"these noble souls dedicate their swords and lives to the battle against evil. Knights, crusaders, and law-bringers, " +
							"paladins seek not just to spread divine justice but to embody the teachings of the virtuous deities they serve.",

		"-Plaguedoctor\n":	"-Rumors abound about the oddly dressed human man on the corner. He stands there calling to those who are sick, " +
							"imploring that he can heal them. Some approach him, and those watching see the man offering small pouches full of paste, or bottles full of oddly-colored liquid. " +
							"Many are curious, but few are brave or desperate enough to approach. One day, a band of adventurers approaches. One of them has succumbed to a debilitating disease.",

		"-Planeswalker\n":	"-Planeswalkers are the source of an infinite energy, called the Planeswalker's Spark. " +
							"This Spark allows them to absorb the mana of a plane, and use it as though it were a tool or weapon. Planeswalkers are able to " +
							"traverse the planes freely, and without restriction; this is called Planeswalking",

		"-Ranger\n":		"-For those who relish the thrill of the hunt, there are only predators and prey. Be they scouts, trackers, or bounty hunters, " +
							"rangers share much in common: unique mastery of specialized weapons, skill at stalking even the most elusive game, " +
							"and the expertise to defeat a wide range of quarries. Knowledgeable, patient, and skilled hunters, these rangers hound man, beast, " +
							"and monster alike, gaining insight into the way of the predator, skill in varied environments, and ever more lethal martial prowess.",

		"-Rogue\n":			"Life is an endless adventure for those who live by their wits. Ever just one step ahead of danger, rogues bank on their cunning, " +
							"skill, and charm to bend fate to their favor. Never knowing what to expect, they prepare for everything, becoming masters of a wide variety of skills, " +
							"training themselves to be adept manipulators, agile acrobats, shadowy stalkers, or masters of any of dozens of other professions or talents.",

		"-Shaman\n":		"-While travelling through swamps, forests and deserts a young human receives directions from the spirits. " +
							"An elf concentrates in its meditation, the spirits around her come forth, to protect her against preying beasts. A half-elf holds one of " +
							"his shamanic focus, his enemies laughs, thinking this is an easy kill, when suddenly from the focus comes a bolt of lightning, " +
							"fulminating the unsuspected enemy Shamans are magic users that gain their powers through spirits. They are the bridge that connects " +
							"the material world and the ethereal world. They can cast spells or use the very spirits to help them in and out of combat, so much, that " +
							"being surrounded by spirits is a common occurrence for them.",

		"-Shaolin\n":		"-Similar to the monk, they are masters of unarmed combat. However, they are less restricted in combat, " +
							"and they sacrifice their Flurry of Blows for increase mobility. In addition, many train with the Arms of Wushu, a set of " +
							"special weapons that are difficult to learn, but devastating to use. They are based out of Shaolin temple in China, where they learn their trade.",

		"-Smuggler\n":		"-Throughout the ages and across civilizations, there has and will always arise the inevitable desire, " +
							"demand, even desperate need, for those things regulated or forbidden. These desires and demands, in turn, frequently spell " +
							"opportunity to an entrepreneurial few. Individuals like you, who—despite oppressive laws, exorbitant taxes, or religious or political " +
							"creeds—have ventured to make available to many what some would see reserved for the few.",

		"-Sorcerer\n":		"-Scions of innately magical bloodlines, the chosen of deities, the spawn of monsters, pawns of fate and destiny, " +
							"or simply flukes of fickle magic, sorcerers look within themselves for arcane prowess and draw forth might few mortals can imagine. " +
							"Emboldened by lives ever threatening to be consumed by their innate powers, these magic-touched souls endlessly indulge in and refine their " +
							"mysterious abilities, gradually learning how to harness their birthright and coax forth ever greater arcane feats.",

		"-Wizard\n":		"-Beyond the veil of the mundane hide the secrets of absolute power. The works of beings beyond mortals, " +
							"the legends of realms where gods and spirits tread, the lore of creations both wondrous and terrible—such mysteries call to " +
							"those with the ambition and the intellect to rise above the common folk to grasp true might. Such is the path of the wizard. " +
							"These shrewd magic-users seek, collect, and covet esoteric knowledge, drawing on cultic arts to work wonders beyond the abilities of mere mortals.",
	}

	return m
}
